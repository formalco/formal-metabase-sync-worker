# Formal Metabase Sync Worker

This is a worker that syncs the formal metabase with the formal database.

## Usage

First build the docker image:

```bash
docker build -t formal-metabase-sync-worker .
```

Then run the docker image:

```bash
docker run -e METABASE_HOSTNAME=""  -e METABASE_USERNAME="" -e METABASE_PASSWORD="" -e METABASE_VERSION="" -e FORMAL_API_KEY="" -e FORMAL_APP_ID="" -e VERIFY_TLS="" formal-metabase-sync-worker 
```

## Environment Variables
- ```METABASE_HOSTNAME```: The hostname of the metabase instance 
- ```METABASE_USERNAME```: The username of the metabase instance
- ```METABASE_PASSWORD```: The password of the metabase instance
- ```METABASE_VERSION```: The version of the metabase instance (e.g.: 0.35.4)
- ```FORMAL_API_KEY```: The API key of the formal instance
- ```FORMAL_APP_ID```: The app id of the Formal Metabase integration
- ```VERIFY_TLS```: Whether or not to verify the TLS certificate of the Metabase instance. Set to `true` or `false`
- ```LOG_LEVEL``` (optional): Set the Global logging level to any of these options: `debug`, `info`, `warn`, `error`, `fatal`, `panic`, `disabled`. Default value is `info`.