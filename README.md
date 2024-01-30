# Formal Metabase Sync Worker

This is a worker that syncs the formal metabase with the formal database.

## Usage

First build the docker image:

```bash
docker build -t formal-metabase-sync-worker .
```

Then run the docker image:

```bash
docker run -e METABASE_HOSTNAME=""  -e METABASE_USERNAME="" -e METABASE_PASSWORD="" -e FORMAL_API_KEY="" -e FORMAL_APP_ID="" formal-metabase-sync-worker
```

## Environment Variables
- ```bash METABASE_HOSTNAME```: The hostname of the metabase instance 
- ```bash METABASE_USERNAME```: The username of the metabase instance
- ```bash METABASE_PASSWORD```: The password of the metabase instance
- ```bash FORMAL_API_KEY```: The API key of the formal instance
- ```bash FORMAL_APP_ID```: The app id of the Formal Metabase integration