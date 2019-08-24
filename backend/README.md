# cbbpoll Backend

## Setting up a Local Development Environment

The easiest way to deploy the cbbpoll backend locally is using
[Docker](https://www.docker.com/products/docker-desktop) and 
[docker-compose](https://docs.docker.com/compose/install/).

To run the backend, `cd` to the `backend/` directory and run (without the $)
```$xslt
$ docker-compose up cbbpoll
```

This will download a Google Datastore Emulator image from Dockerhub,
create a volume to persist application data, spin up a container for the
Datastore image, build an image for the `cbbpoll` backend, and finally
spin up a container for this image.  At this point, the `cbbpoll` server
should be up and listening for requests on port 8000 of your local machine.

```$xslt
$ curl localhost:8000/ping
  {"Version":"0.1"} # output
```

To rebuild the backend after making changes to the source code:

```$xslt
$ docker-compose build cbbpoll
```

This should rebuild the image from source, while caching dependencies.
Bring the image up again as above to continue making requests against it.

The docker-compose file creates a [Docker volume](https://docs.docker.com/storage/volumes/)
for the purpose of persisting the application data across executions.

To wipe this data clean and start from scratch, run:

```$xslt
$ docker-compose down -v
```

This will bring down any of the project's running containers and remove
the data volume.

## API Documentation
todo

## Import of Database Data
todo
