# Formal BI Sync Worker

A worker that syncs users from BI tools (Metabase, Omni) to Formal by mapping user identities via external IDs. Configure one or both integrations — any integration with its required env vars set will be enabled.

## Usage

First build the docker image:

```bash
docker build -t formal-bi-sync-worker .
```

Then run the docker image:

```bash
docker run \
  -e FORMAL_API_KEY="" \
  -e VERIFY_TLS="" \
  -e LOG_LEVEL="" \
  -e FREQUENCY="" \
  -e METABASE_HOSTNAME="" \
  -e METABASE_BI_INTEGRATION_ID="" \
  -e METABASE_USE_API_KEY="" \
  -e METABASE_API_KEY="" \
  -e METABASE_USERNAME="" \
  -e METABASE_PASSWORD="" \
  -e METABASE_VERSION="" \
  -e CF_ACCESS_CLIENT_ID="" \
  -e CF_ACCESS_CLIENT_SECRET="" \
  -e OMNI_API_KEY="" \
  -e OMNI_HOSTNAME="" \
  -e OMNI_BI_INTEGRATION_ID="" \
  formal-bi-sync-worker
```

## Environment Variables

### Formal (required)
- `FORMAL_API_KEY`: API key for the Formal instance
- `VERIFY_TLS`: Whether to verify TLS certificates. Set to `true` or `false`
- `LOG_LEVEL` (optional): Global log level — `debug`, `info`, `warn`, `error`, `fatal`, `panic`, `disabled`. Defaults to `info`
- `FREQUENCY` (optional): How often to run the sync (e.g. `1h`, `30m`). If not set, runs once and exits

### Metabase (optional — enabled when `METABASE_HOSTNAME` is set)
- `METABASE_HOSTNAME`: Hostname of the Metabase instance
- `METABASE_BI_INTEGRATION_ID`: Integration ID of the Formal Metabase integration
- `METABASE_USE_API_KEY` (optional): Use API key auth instead of username/password. Set to `true` or `false`. Defaults to `false`
- `METABASE_API_KEY` (optional): Metabase API key; required if `METABASE_USE_API_KEY` is `true`
- `METABASE_USERNAME` (optional): Metabase username; required if not using API key auth
- `METABASE_PASSWORD` (optional): Metabase password; required if not using API key auth
- `METABASE_VERSION`: Version of the Metabase instance (e.g. `0.35.4`)
- `CF_ACCESS_CLIENT_ID` (optional): Cloudflare Access Client ID for Metabase instances behind Cloudflare Access
- `CF_ACCESS_CLIENT_SECRET` (optional): Cloudflare Access Client Secret for Metabase instances behind Cloudflare Access

### Omni (optional — enabled when both `OMNI_API_KEY` and `OMNI_HOSTNAME` are set)
- `OMNI_API_KEY`: Omni organization API key
- `OMNI_HOSTNAME`: Hostname of the Omni instance (e.g. `myorg.omniapp.co`)
- `OMNI_BI_INTEGRATION_ID`: Integration ID of the Formal Omni BI integration
