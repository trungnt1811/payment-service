#inspect_args
output_directory=$(readlink -f "${args[--directory]}")

if [[ ! -z "${args[--filename]}" ]]; then
  output_filename="${args[--filename]}"
  else
    output_filename="life-ai.life-point.backup.$(date +%Y%m%d-%H%M%S)"
fi

if ! [[ -d "${output_directory}" ]]; then
  echo "Create backup directory..."
  mkdir -p "${output_directory}"
fi

(
  echo "Create backup zip archive..."
  cd "${root_directory}" && \
  zip \
    -r "${output_directory}/${output_filename}.zip" \
    . \
    -1 \
    -x \
      \*backups\* \
)

echo "Backup has been created successfully. Location: ${output_directory}/${output_filename}.zip"
