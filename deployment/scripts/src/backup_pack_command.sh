#inspect_args
output_directory=$(readlink -f "${args[--directory]}")

if [[ ! -z "${args[--filename]}" ]]; then
  output_filename="${args[--filename]}"
  else
    output_filename="life-ai.life-point.deployment.$(date +%Y%m%d-%H%M%S)"
fi

if ! [[ -d "${output_directory}" ]]; then
  echo "Create backup directory..."
  mkdir -p "${output_directory}"
fi

declare -a exclude_patterns=(
  \*archives\*
  \*backups\*
  \*scripts/src\*
  .gitignore
#  .env
  .env.\*
)

if ! [[ ${args[--include-env]} == 1 ]]; then
  exclude_patterns+=(.env)
fi

(
  echo "Create deployment zip archive..."
  cd "${root_directory}" && \
  zip \
    -r "${output_directory}/${output_filename}.zip" \
    . \
    -1 \
    -x "${exclude_patterns[@]}"
)

echo "Deployment package has been created successfully. Location: ${output_directory}/${output_filename}.zip"
