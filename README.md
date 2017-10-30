# go-bandit-server

A server running multi-armed-bandit algorithm. A redis server is required to implement caching and delayed learning.

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
