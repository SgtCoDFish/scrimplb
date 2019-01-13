#!/usr/bin/env python3

import sys
import boto3
import botocore
import subprocess

from typing import List, Tuple


CONFIG_STR = """upstream {application} {{
{servers}
    server [::1]:9090 backup;
}}"""

UPSTREAM_LOCATION = "/etc/sgtcodfish/upstream.conf"


def load_bucket_name() -> str:
    with open("/etc/sgtcodfish/lb-bucket-name", "r") as f:
        bucket_name = f.read()

    return bucket_name


def load_applications() -> List[str]:
    with open("/etc/sgtcodfish/lb-applications", "r") as f:
        applications = f.read().split(",")

    return applications


def fetch_application_lbs(s3: botocore.client.BaseClient,
                          bucket_name: str,
                          application: str) -> List[Tuple[str, str]]:
    try:
        response = s3.get_object(Bucket=bucket_name, Key=application)
    except botocore.exceptions.ClientError:
        print("Failed to fetch", application, "from S3")
        raise

    lines = [l.split(" ") for l in sorted(response["Body"].read().split("\n"))]

    return [(lb[0], lb[1]) for lb in lines]


def current_upstream() -> str:
    with open(UPSTREAM_LOCATION, "r") as f:
        contents = f.read()

    return contents.strip()


def write_new_config(config: str) -> None:
    with open(UPSTREAM_LOCATION, "w") as f:
        f.write(UPSTREAM_LOCATION)


def reload_nginx():
    subprocess.run(["systemctl", "reload", "nginx"])


def main() -> None:
    s3 = boto3.client("s3")
    bucket_name = load_bucket_name()
    applications = load_applications()

    total_config = ""

    for application in applications:
        lbs = fetch_application_lbs(s3, bucket_name, application)

        nginx_config = CONFIG_STR.format(
            application=application,
            servers="\n".join([
                "    server [{}]:{};".format(lb[0], lb[1])
            ] for lb in lbs)
        )

        total_config += nginx_config + "\n"

    total_config = total_config.strip()

    upstream = current_upstream()

    if upstream == total_config:
        print("Upstream matches generated config, nothing to do")
        sys.exit(0)

    write_new_config(total_config)
    reload_nginx()
    print("Wrote new config")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print("Fatal error:", e)
        sys.exit(1)
