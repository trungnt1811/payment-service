#inspect_args
full_output_directory=$(readlink -f "${args[--directory]}")
(
  cd "$root_directory"
  mkdir -p "${full_output_directory}"
  zip -r "${full_output_directory}/${args[--filename]}.zip" . \
      -1 \
      -x \
        \*.git\* \*.idea\* \*archives\* \*cloudformation\* .env.\* README.md \
)
