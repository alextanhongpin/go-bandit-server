# go-bandit-server

A server running multi-armed-bandit algorithm. A redis server is required to implement caching and delayed learning.

![graph](assets/graph.png)

<!--
## Redis Commands

```bash
# Get all keys with the arm identifier
$ KEYS arm:*

# Flush
$ FLUSHALL

# Get all hash keys and values
$ HGETALL key

# Get by timestamp
$ ZRANGE arm 0 -1 WITHSCORES
```
-->


## TODO

- [ ] dockerize the service
- [ ] create a dashboard for the service (web ui)
- [ ] make the service configurable
- [ ] make the results of the service transparent (like feature X is n% better than feature Y)
- [ ] isolate configuration
- [ ] add labels and annotations for registry
- [ ] create sdk to integrate the functionality into other applications
- [ ] or make this into a sidecar proxy
- [ ] CRUD and API interface and CLI to integrate with the configuration
- [ ] fix the race issue that is found in the api
