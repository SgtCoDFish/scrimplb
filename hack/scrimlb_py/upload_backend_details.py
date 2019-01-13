import boto3
import botocore


class Application:

    def __init__(self, name, port) -> None:
        self.name = name
        self.port = port


def load_bucket_name() -> str:
    with open("/etc/sgtcodfish/lb-bucket-name", "r") as f:
        bucket_name = f.read().strip()

    return bucket_name


def load_application() -> Application:
    with open("/etc/sgtcodfish/application", "r") as f:
        raw_app = f.read().strip()

    vals = raw_app.split(" ")
    app = Application(vals[0], vals[1])
    return app


def fetch_current_lb_config(s3: botocore.client.BaseClient,
                            bucket_name: str,
                            application: Application) -> str:
    try:
        response = s3.get_object(Bucket=bucket_name, Key=application)
    except botocore.exceptions.ClientError:
        print(
            "Failed to fetch currently registered upstreams for", application
        )
        raise

    lines = [l.split(" ") for l in sorted(response["Body"].read().split("\n"))]
    return [(lb[0], lb[1]) for lb in lines]


def main() -> None:
    s3 = boto3.client("s3")

    bucket_name = load_bucket_name()
    application = load_application()
    current_config = fetch_current_lb_config(s3, bucket_name, application)

    print(current_config)
