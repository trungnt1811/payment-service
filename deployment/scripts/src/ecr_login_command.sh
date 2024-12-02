inspect_args

.env --file ${root_directory}/.env

# Detect instance type
# Auto-detect instance type
if curl -s "http://metadata.google.internal" -H "Metadata-Flavor: Google" > /dev/null 2>&1; then
    INSTANCE_TYPE="GCP"
elif curl -s "http://169.254.169.254/latest/meta-data/" > /dev/null 2>&1; then
    INSTANCE_TYPE="AWS"
else
    echo "Instance type is not GCP nor AWS"
fi

# Conditional logic based on instance type
if [ "$INSTANCE_TYPE" = "GCP" ]; then
    echo "Running on GCP Compute Engine"
    # Logic specific to GCP
    AWS_ROLE_ARN=$(.env get AWS_ROLE_ARN && echo "${REPLY}")
    TOKEN_FILE="${TOKEN_FILE:-/var/run/secrets/google/tokens}"

    # Check if TOKEN_FILE exists and expiration timestamp
    if [ ! -f "$TOKEN_FILE" ]; then
        echo "Fetching new token for GCP..."
        sudo curl -H "Metadata-Flavor: Google" \
            "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=$AWS_ROLE_ARN" \
            -o "$TOKEN_FILE"
    else
        EXPIRATION=$(cat "$TOKEN_FILE" | cut -d "." -f2 | base64 -d | jq -r '.exp' 2>/dev/null)
        if [ -z "$EXPIRATION" ] || ! [[ $EXPIRATION =~ ^[0-9]+$ ]] || [ "$(date +%s)" -ge "$EXPIRATION" ]; then
            echo "Fetching new token for GCP..."
            sudo curl -H "Metadata-Flavor: Google" \
                "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=$AWS_ROLE_ARN" \
                -o "$TOKEN_FILE"
        fi
    fi

elif [ "$INSTANCE_TYPE" = "AWS" ]; then
    echo "Running on AWS EC2"
    # Logic specific to AWS
    echo "Assuming role directly using instance metadata..."
    # AWS EC2-specific setup (assumes IAM role attached to the instance)
fi

# Retrieve AWS credentials
CREDENTIALS=$(aws sts assume-role-with-web-identity \
    --role-arn $AWS_ROLE_ARN \
    --role-session-name portable-session \
    --web-identity-token file://$TOKEN_FILE \
    --query 'Credentials' --output json)

export AWS_ACCESS_KEY_ID=$(echo $CREDENTIALS | jq -r '.AccessKeyId')
export AWS_SECRET_ACCESS_KEY=$(echo $CREDENTIALS | jq -r '.SecretAccessKey')
export AWS_SESSION_TOKEN=$(echo $CREDENTIALS | jq -r '.SessionToken')

if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    echo "Failed to retrieve AWS credentials."
    exit 1
fi

aws --profile="${args[--profile]}" ecr get-login-password --region ap-southeast-1 | docker login --username AWS --password-stdin "${args[url]}"
