# API First

## Async API
https://www.asyncapi.com/docs/tools/cli/usage

```
asyncapi validate api-specification/asyncapi.yaml
```

## Docker
https://hub.docker.com/r/asyncapi/

https://hub.docker.com/r/asyncapi/generator

````
docker run --rm -it \
--user=root \
-v api-specification/asyncapi.yaml:/app/asyncapi.yml \
-v build:/app/output \
asyncapi/cli --help
````

````
docker run --rm -it \
--entrypoint='' \
--user=root \
-v ./api-specification/asyncapi.yml:/app/asyncapi.yml \
-v ./build:/app/output \
asyncapi/cli:3.1.1 \
asyncapi validate /app/asyncapi.yml
````