#inspect_args

echo "${args[services]}"

if [[ -z "${args[services]}" ]]
then
  echo "# Start all services..."
else
  echo "# Start ${args[services]} services..."
fi

(cd ${directory}/.. && docker compose up -d ${args[services]//,/ })
