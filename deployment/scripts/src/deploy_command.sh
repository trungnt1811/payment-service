#inspect_args

(
  cd "${directory}" \
  && ./cli ecr login \
  && ./cli service build \
  && ./cli service start \
  && ./cli service init \
  && ./cli service cleanup
)
#
