# Event Feed Engine

Event-Driven News Feed Backend built in Go.

## Objective / Problem definition
Design and implement the backend service that: 
- Accept user generated events(posts)
- Propagates these events to relevant users(followers)
- Maintains the personalized feed for each users
- Serves the feed efficiently with low latency 
- Protects API from overload and abuse 
- Scales horizontal as usuage grows 
The system should be event-driven, concurrent and cache-aware, reflecting real-world backend architectures used by large-scale applications.

## Functional requirements
  1. User Management - authentication using jwt token and every user is identified using unique identifier.

  2. Social - user can follow and unfollow the users and maintain relationships.

  3. Post creation - user can create post communicating their insights, thoughts, etc and the format can be text, media, etc.

  4. Create the Feed - the feed is created and stored in the cache.

  5. Read post / Retrieve the feed - user should be able to view the feed when opened. 

  6. See the users profile and posts - every user can see the profile and posts of the other users 

  7. Rate limiting - API must enforce limits 

## Non-Functional requirements
  1. Scalability - can be horizontally scalable.
  2. Performance - low latency feed retrieval.
  3. Consistency Model - moderate delay is acceptable.
  4. Reliability - avoid duplicate feeds.
  5. Security - rate liminting and authentication.
  6. Extensibility - system supports future additions.

## System design 
![alt text](<README_IMAGES/systemDesign.png>)

## Database Schema
![alt text](<README_IMAGES\databaseSchema.png>)

This project is under active development.