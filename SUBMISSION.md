# Submission

## Enhancements

- The `db` code now includes a `context`  
- Gateways are selected based off of their currency and country. They are randomly selected when multiple are found in order to spread the load. Similarly you could use a `weighted` attribute which specifies preference.
- Middleware with generics is added to handle correct typing and parsing of data in the body. This allows the handlers to be more concentrated around handling the request rather than parsing it.
- For fault tolerance there is an incremental back off with retries which calls the circuit breaker as part of a hybrid approach. 
- Security wise I've added AES cryptography which encrypts fields within json/xml messages, all except the gateway id as this is insensitive and a gateway consumer would need to read this to see if it should be picked up or not.
- Kafka UI, I've added kafka UI into the compose file so you can see the messages being pushed.
- DB singleton, it was already set up this way but a GetDB() method that gets the DB instance has been added
- DB transactions are used so that only on full success are writes committed, otherwise they're rolled back.

## Out of Scope

- There are amount validations but very basic. Realistically you'd have minimum withdrawal/deposit amounts that would be set as config
- Swagger docs, sorry, would simply take too much time.
- Logging. It was my intention to build a verbose and customizable logger for this exercise. But again, due to time constraints, this hasn't been done.
- Extensive tests. Again, lack of time. For very sensitive functions like this I'd add `fuzzing` to the test suite as well as `benchmarks`

## Notes

- Deposit and Withdrawal are almost identical in their implementations. Depending on enhancements they could potentially be separate endpoints with the same handler. It's usually better to start separate and then merge later. Further requirements gathering is needed

## Running And Test

I've used docker-compose `merge` files. You have the main `docker-compose.yml` file and then a `docker-compose.dev.yml` which incorporates hot reloading with `air` for go. Testing is also done in a similar way.

For hot reloading you can do `docker-compose -f docker-compose.yml -f docker-compose.dev.yml up` or just run the `./dev.sh` file.

For testing you can do similar with `docker-compose -f docker-compose.yml -f docker-compose.test.yml up app` or `./test.sh`.

Rebuild is needed when moving between environments e.g. `docker-compose -f docker-compose.yml -f docker-compose.dev.yml build` when requiring live reloading and `docker-compose build` when wanting the production like server.

Testing incorporates the other `docker-compose` services into the test suite.

Testing additionally has not been extensive. I did a partial test with `sqlmock` in the `handlers_test.go` file so you could see the idea. It would simply take too much time to flesh out everything. 

Honestly, there's a lot of stuff I would add into this, however, I'm acutely aware the context of this is an assignment. There's lots of ways the code could fail and so more advanced fault tolerance, better testing, heavy refactoring (you need time to break everything down further and unit test) to name a few.