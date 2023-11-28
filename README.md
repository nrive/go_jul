# go_jul

## Files
- config/users.csv  
CSV file containing users, username is used to identify the con and won files for the spesific users  
Format: username,password  
Default: demo/demo  
- config/username_con.csv  
CSV file containing contestants  
Format: user,default_selected  
- config/username_won.csv  
CSV file containing winners  

## Build go_jul
Build go_jul binary for use with docker
```
$ GOOS=linux GOARCH=amd64 go build -o ../app/go_jul go_jul.go
```

## Docker directory structure
- go_jul
  - config
  - app
    - static
    - templates

## Build docker image
Dockerfile used to build the docker image
```
FROM alpine:latest
RUN addgroup -g 1000 -S notroot && adduser -u 1000 -S notroot -G notroot

FROM scratch
COPY --from=0 /etc/group /etc/passwd /etc/
USER notroot:notroot
COPY ./app /
EXPOSE 9000
ENTRYPOINT ["/go_jul"]
```

## Run docker container
Docker compose file used to run the container
```
version: '3.5'
services:
  go_jul:
    build: go_jul
    container_name: go_jul
    restart: unless-stopped
    volumes:
      - "/data/docker/go_jul/config:/config"
    networks:
      - exposed
```
