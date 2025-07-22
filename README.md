# Formal Metabase Sync Worker

This is a worker that syncs the formal metabase with the formal database.

## Usage

First build the docker image:

```bash
docker build -t formal-metabase-sync-worker .
```

Then run the docker image:

```bash
docker run -e METABASE_HOSTNAME="" -e METABASE_USE_API_KEY="" -e METABASE_API_KEY=""  -e METABASE_USERNAME="" -e METABASE_PASSWORD="" -e METABASE_VERSION="" -e FORMAL_API_KEY="" -e FORMAL_APP_ID="" -e VERIFY_TLS="" -e CF_ACCESS_CLIENT_ID="" -e CF_ACCESS_CLIENT_SECRET="" formal-metabase-sync-worker 
```

## Environment Variables
- ```METABASE_USE_API_KEY``` (optional): Whether or not to use the Metabase API key instead of username/password auth. Set to `true` or `false`. Default value is `false`.
- ```METABASE_API_KEY``` (optional): The API key for the metabase instance; required if not providing username/password auth.
- ```METABASE_HOSTNAME```: The hostname of the metabase instance 
- ```METABASE_USERNAME```: (optional) The username of the metabase instance; required if not providing API key.
- ```METABASE_PASSWORD```: (optional) The password of the metabase instance; required if not providing API key.
- ```METABASE_VERSION```: The version of the metabase instance (e.g.: 0.35.4)
- ```FORMAL_API_KEY```: The API key of the formal instance
- ```FORMAL_APP_ID```: The app id of the Formal Metabase integration
- ```VERIFY_TLS```: Whether or not to verify the TLS certificate of the Metabase instance. Set to `true` or `false`
- ```CF_ACCESS_CLIENT_ID``` (optional): Cloudflare Access Client ID for bypassing Cloudflare Access protection on Metabase instances
- ```CF_ACCESS_CLIENT_SECRET``` (optional): Cloudflare Access Client Secret for bypassing Cloudflare Access protection on Metabase instances
- ```LOG_LEVEL``` (optional): Set the Global logging level to any of these options: `debug`, `info`, `warn`, `error`, `fatal`, `panic`, `disabled`. Default value is `info`.
- ```FREQUENCY``` (optional): The frequency at which to run the sync. Expected format: `1h`, `30m`, etc. If not provided, the sync will run once and then exit.
