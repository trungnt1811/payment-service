# Code here runs inside the initialize() function
# Use it for anything that you need to run before any other function, like
# setting environment vairables:
# CONFIG_FILE=settings.ini
#
# Feel free to empty (but not delete) this file.

current_directory=$PWD
directory="$(cd "$(dirname "$0")" && pwd)"
root_directory=$(readlink -f "${directory}/..")

source ${directory}/helpers/dotenv
