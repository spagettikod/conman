## Build
```
docker build -t spagettikod/conman .
```

## Development run
```
docker build -t spagettikod/conman . && docker run --rm -p 8080:8080 -v $(pwd)/www:/www -v /var/run/docker.sock:/var/run/docker.sock spagettikod/conman
```

## Production run
```
docker run --rm -p 26652:8080 -v /var/run/docker.sock:/var/run/docker.sock spagettikod/conman
```