#inspect_args
(
  cd "${root_directory}" && \
  docker network inspect nginx-proxy --format {{.Id}} 2>/dev/null || docker network create nginx-proxy
)
