#inspect_args

echo "# Build the services..."
(cd ${directory}/.. && docker compose build ${args[services]})
