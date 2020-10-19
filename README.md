# Build image
```
docker build -t <user>/<repo>:<tag> .
```
Cross platform (rpi):
```
docker buildx build --platform linux/arm/v7 -t <user>/<repo>:<tag> --push .
``` 

# Usage

### Docker compose

```
...

  backup-scheduler:
    image: <user>/<repo>:<tag>
    restart: unless-stopped
    volumes: 
      - /path_to_file1:/files/file1
      - /path_to_file2:/files/file2
      - /path_to_config/prod.yaml:/app/config/prod.yaml
    environment:
      - ENV=prod

...
```
