#inspect_args

echo "# Update the services..."
(cd ${directory}/.. && docker compose pull ${args[services]})
