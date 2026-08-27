# Docker troubleshooting bundle

A useful Docker bundle contains bounded container logs, `docker inspect` JSON, Compose configuration with secrets removed, image metadata, and daemon diagnostics relevant to the incident window.

Never include registry credentials, raw `.env` files, private keys, or mounted secret contents. The analyzer treats inspect output as data and never executes image entrypoints or bundle scripts.

Compare expected and observed health status, restart count, exit code, memory limits, mounts, networks, and timestamps. An exit code or OOM flag is evidence about termination, while the application logs usually provide the next diagnostic step.

The current built-in pipeline provides generic JSON/config/log analysis for Docker artifacts. Confirm any container-specific interpretation using `docker inspect`, `docker logs --since`, and the daemon event history on the source host.
