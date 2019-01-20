package seed

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Provider can retrieve a seed from an object in an S3 bucket
type S3Provider struct {
	Bucket           string
	AvailabilityZone string
	Region           string

	// accessKey       string
	// secretAccessKey string
}

// NewS3Provider creates a new S3 seed provider from the given config.
// "availability-zone" is optional; if not given it will be deduced from instance
// metadata if running on EC2.
func NewS3Provider(config map[string]string) (*S3Provider, error) {
	required := [...]string{"bucket", "region", "availability-zone"}

	for _, r := range required {
		_, ok := config[r]

		if !ok {
			return nil, fmt.Errorf("missing required s3 provider config: %s", r)
		}
	}

	bucket := config["bucket"]
	region := config["region"]
	az := config["availability-zone"]

	return &S3Provider{
		Bucket:           bucket,
		AvailabilityZone: az,
		Region:           region,
	}, nil
}

// FetchSeed makes a call to S3 to retrieve the seed for the AZ this
// instance is located in.
func (s *S3Provider) FetchSeed() (Seeds, error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(s.Region)},
	)
	downloader := s3manager.NewDownloader(sess)

	buf := aws.NewWriteAtBuffer([]byte{})

	numBytes, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(s.AvailabilityZone),
		})

	if err != nil {
		return Seeds{}, fmt.Errorf("unable to download seed %v", err)
	}

	fmt.Println("Downloaded", numBytes, "bytes")

	return Seeds{
		Seeds: []Seed{
			{
				strings.TrimSpace(string(buf.Bytes())),
				9999,
			},
		},
	}, nil
}

// func deduceInstanceAZ() (string, error) {
// 	resp, err := http.Get("http://169.254.169.254/latest/meta-data/placement/availability-zone")
//
// 	if err != nil {
// 		return "", nil
// 	}
//
// 	defer resp.Body.Close()
// 	raw, err := ioutil.ReadAll(resp.Body)
//
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return string(raw), nil
// }
