import abc
import sys
import boto3
import botocore


class Configurator:

    def __init__(self, application_name: str) -> None:
        self.application_name = application_name

    def load_cache_key(self):
        filename = "/etc/sgtcodfish/{}.cachekey".format(self.application_name)
        with open(filename, "r") as f:
            contents = f.read()

        self.cache_key = contents

    @abc.abstractmethod
    def fetch_configuration(self):
        pass

    @abc.abstractmethod
    def should_upload(self) -> bool:
        pass

    @abc.abstractmethod
    def apply_configuration(self):
        pass


class S3Configurator(Configurator):

    def __init__(self, application_name: str, bucket_name: str) -> None:
        super().__init__(application_name)
        self.bucket_name = bucket_name
        self.s3 = boto3.client("s3")

        self.load_cache_key()

    def head_configuration(self):
        try:
            response = self.s3.get_object(
                Bucket=self.bucket_name,
                Key=self.application,
                IfNoneMatch=self.cache_key
            )
        except botocore.exceptions.ClientError as e:
            print("Couldn't fetch config:", e, file=sys.stderr)

        print(response)
