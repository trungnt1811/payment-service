#inspect_args
output_directory=$(readlink -f "${args[--directory]}")
keep=${args[--keep]}
(
  # shellcheck disable=SC2164
  cd "${output_directory}";
  echo "Remove old backup archives, only keep ${keep} last ones..."
  echo "## The following archives will be removed:"
  # shellcheck disable=SC2012
  ls -dt ./*.zip | tail -n +"$((${keep} + 1))"
#  rm -f "$(ls -dt *.zip)" | tail -n +"$((${keep} + 1))"
  # shellcheck disable=SC2012
  echo "## Remaining archives:"
  ls -dt ./*.zip | tail -n -"${keep}"
  echo "## Start cleaning up..."
  rm -f $(ls -dt ./*.zip | tail -n +"$((${keep} + 1))");
  ## TODO: Ask for confirmation (yes/no)
  echo "## Backup archives has been cleaned up!"
)
