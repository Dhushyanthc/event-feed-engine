param(
    [string]$BaseUrl = "http://localhost:8081",
    [switch]$SkipDbQueueCheck
)

$ErrorActionPreference = "Stop"

function To-JsonBody($Value) {
    $Value | ConvertTo-Json -Compress
}

function Assert($Condition, $Message) {
    if (-not $Condition) {
        throw $Message
    }
}

function Invoke-ApiJson {
    param(
        [string]$Method,
        [string]$Uri,
        [hashtable]$Headers,
        [object]$Body
    )

    $params = @{
        Method = $Method
        Uri = $Uri
    }

    if ($Headers) {
        $params.Headers = $Headers
    }

    if ($null -ne $Body) {
        $params.ContentType = "application/json"
        $params.Body = (To-JsonBody $Body)
    }

    Invoke-RestMethod @params
}

function Invoke-ExpectedStatus {
    param(
        [string]$Method,
        [string]$Uri,
        [int]$ExpectedStatus,
        [hashtable]$Headers,
        [object]$Body,
        [Microsoft.PowerShell.Commands.WebRequestSession]$WebSession
    )

    $params = @{
        Method = $Method
        Uri = $Uri
        UseBasicParsing = $true
        ErrorAction = "Stop"
    }

    if ($Headers) {
        $params.Headers = $Headers
    }

    if ($WebSession) {
        $params.WebSession = $WebSession
    }

    if ($null -ne $Body) {
        $params.ContentType = "application/json"
        $params.Body = (To-JsonBody $Body)
    }

    try {
        $response = Invoke-WebRequest @params
        if ($response.StatusCode -ne $ExpectedStatus) {
            throw "Expected status $ExpectedStatus from $Method $Uri but got $($response.StatusCode)."
        }
        return $response
    }
    catch {
        $status = $_.Exception.Response.StatusCode.value__
        if ($status -ne $ExpectedStatus) {
            throw "Expected status $ExpectedStatus from $Method $Uri but got $status."
        }
        return $_.Exception.Response
    }
}

function Invoke-PostgresScalar {
    param(
        [string]$Sql
    )

    $output = docker compose exec -T postgres psql `
        -U postgres `
        -d event_feed_engine `
        -t `
        -A `
        -F "," `
        -c $Sql

    if ($LASTEXITCODE -ne 0) {
        throw "PostgreSQL query failed."
    }

    $line = $output | Where-Object { $_ -and $_.Trim() -ne "" } | Select-Object -First 1
    if (-not $line) {
        return $null
    }

    return $line.Trim()
}

$stamp = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
$password = "pass1234"
$wrongPassword = "wrong-pass"

$aliceEmail = "alice+$stamp@example.com"
$bobEmail = "bob+$stamp@example.com"

Write-Host "Using API:" $BaseUrl

$alice = Invoke-ApiJson -Method Post -Uri "$BaseUrl/users" -Body @{
    name = "Alice"
    email = $aliceEmail
    password = $password
}

$bob = Invoke-ApiJson -Method Post -Uri "$BaseUrl/users" -Body @{
    name = "Bob"
    email = $bobEmail
    password = $password
}

$aliceLogin = Invoke-ApiJson -Method Post -Uri "$BaseUrl/login" -Body @{
    email = $aliceEmail
    password = $password
}

$bobLogin = Invoke-ApiJson -Method Post -Uri "$BaseUrl/login" -Body @{
    email = $bobEmail
    password = $password
}

$aliceHeaders = @{
    Authorization = "Bearer $($aliceLogin.token)"
}

$bobHeaders = @{
    Authorization = "Bearer $($bobLogin.token)"
}

$follow = Invoke-ApiJson -Method Post -Uri "$BaseUrl/follow" -Headers $aliceHeaders -Body @{
    following_id = $bob.id
}

$postContent = "smoke-test-post-$stamp"

$post = Invoke-ApiJson -Method Post -Uri "$BaseUrl/posts" -Headers $bobHeaders -Body @{
    content = $postContent
    media_url = ""
}

$feed = $null
$feedPosts = @()
$match = $null

for ($attempt = 1; $attempt -le 20; $attempt++) {
    $feed = Invoke-ApiJson -Method Get -Uri "$BaseUrl/feed" -Headers $aliceHeaders

    $feedPosts = @($feed.posts)
    $match = $feedPosts | Where-Object { $_.id -eq $post.id }

    if ($match) {
        break
    }

    Start-Sleep -Milliseconds 500
}

$feedEventState = $null
if (-not $SkipDbQueueCheck) {
    for ($attempt = 1; $attempt -le 20; $attempt++) {
        $feedEventState = Invoke-PostgresScalar -Sql @"
SELECT processed, failed, attempts
FROM feed_events
WHERE post_id = $($post.id)
  AND user_id = $($bob.id)
ORDER BY id DESC
LIMIT 1;
"@

        if ($feedEventState -and $feedEventState.StartsWith("t,")) {
            break
        }

        Start-Sleep -Milliseconds 500
    }
}

$unauthorizedFeed = Invoke-ExpectedStatus -Method Get -Uri "$BaseUrl/feed" -ExpectedStatus 401
$wrongPasswordLogin = Invoke-ExpectedStatus -Method Post -Uri "$BaseUrl/login" -ExpectedStatus 401 -Body @{
    email = $aliceEmail
    password = $wrongPassword
}
$selfFollow = Invoke-ExpectedStatus -Method Post -Uri "$BaseUrl/follow" -ExpectedStatus 400 -Headers $aliceHeaders -Body @{
    following_id = $alice.id
}
$emptyPost = Invoke-ExpectedStatus -Method Post -Uri "$BaseUrl/posts" -ExpectedStatus 400 -Headers $bobHeaders -Body @{
    content = ""
    media_url = ""
}

$rateLimited = $false
$loginSession = New-Object Microsoft.PowerShell.Commands.WebRequestSession
for ($attempt = 1; $attempt -le 12; $attempt++) {
    $response = Invoke-ExpectedStatus -Method Post -Uri "$BaseUrl/login" -ExpectedStatus 401 -Body @{
        email = $aliceEmail
        password = $wrongPassword
    } -WebSession $loginSession

    if ($response.StatusCode -eq 429) {
        $rateLimited = $true
        break
    }

    try {
        Invoke-WebRequest -Method Post -Uri "$BaseUrl/login" -WebSession $loginSession -UseBasicParsing -ErrorAction Stop -ContentType "application/json" -Body (To-JsonBody @{
            email = $aliceEmail
            password = $wrongPassword
        }) | Out-Null
    }
    catch {
        if ($_.Exception.Response.StatusCode.value__ -eq 429) {
            $rateLimited = $true
            break
        }
    }
}

Assert ($alice.id -gt 0) "User creation failed for Alice."
Assert ($bob.id -gt 0) "User creation failed for Bob."
Assert ([bool]$aliceLogin.token) "Alice login did not return a token."
Assert ([bool]$bobLogin.token) "Bob login did not return a token."
Assert ($follow.following_id -eq $bob.id) "Follow request did not target Bob."
Assert ($post.content -eq $postContent) "Created post content did not match the request."
Assert ([bool]$match) "Smoke test failed: created post was not found in follower feed."
if (-not $SkipDbQueueCheck) {
    Assert ([bool]$feedEventState) "Smoke test failed: no feed_events row was found for the created post."
    Assert ($feedEventState.StartsWith("t,")) "Smoke test failed: feed_events row was not processed. State: $feedEventState"
}
Assert ($unauthorizedFeed.StatusCode -eq 401) "Expected /feed without auth to return 401."
Assert ($wrongPasswordLogin.StatusCode -eq 401) "Expected /login with wrong password to return 401."
Assert ($selfFollow.StatusCode -eq 400) "Expected self-follow to return 400."
Assert ($emptyPost.StatusCode -eq 400) "Expected empty post creation to return 400."

[pscustomobject]@{
    alice = $alice
    bob = $bob
    follow = $follow
    post = $post
    feed_post_count = $feedPosts.Count
    post_found_in_feed = [bool]$match
    feed_event_state = $feedEventState
    unauthorized_feed_status = $unauthorizedFeed.StatusCode
    wrong_password_status = $wrongPasswordLogin.StatusCode
    self_follow_status = $selfFollow.StatusCode
    empty_post_status = $emptyPost.StatusCode
    rate_limit_triggered = $rateLimited
} | ConvertTo-Json -Depth 6

if (-not $rateLimited) {
    Write-Warning "Rate limit test did not trigger 429. This likely means the current limiter keying is too weak for unauthenticated requests."
}
