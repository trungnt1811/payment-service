#inspect_args

echo "${args[services]}"

if [[ -z "${args[services]}" ]]
then
  echo "# Stop all services..."
  (cd ${directory}/.. && docker compose down)
else
  echo "# Stop ${args[services]} services..."
  (cd ${directory}/.. && docker compose stop ${args[services]})
fi
