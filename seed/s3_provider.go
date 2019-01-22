package seed

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sgtcodfish/scrimplb/constants"
)

// S3Provider can retrieve a seed from an object in an S3 bucket
type S3Provider struct {
	Bucket           string
	AvailabilityZone string
	Region           string
	Key              string

	// accessKey       string
	// secretAccessKey string
}

// NewS3Provider creates a new S3 seed provider from the given config.
// "availability-zone" is optional; if not given it will be deduced from instance
// metadata if running on EC2.
func NewS3Provider(config map[string]interface{}) (*S3Provider, error) {
	var provider S3Provider

	err := mapstructure.Decode(config, &provider)

	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse s3 provider config")
	}

	if provider.Bucket == "" {
		return nil, errors.New("missing required 'bucket' in provider config")
	}

	if provider.Region == "" {
		return nil, errors.New("missing required 'region' in provider config")
	}

	if provider.Key == "" {
		provider.Key = constants.DefaultKey
	}

	return &provider, nil
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
			Key:    aws.String(s.Key),
		})

	if err != nil {
		return Seeds{}, fmt.Errorf("unable to download seed %v", err)
	}

	fmt.Println("Downloaded", numBytes, "bytes")

	return Seeds{
		Seeds: []Seed{
			{
				strings.TrimSpace(string(buf.Bytes())),
				"9999",
			},
		},
	}, nil
}

func (s *S3Provider) PushSeed() error {
	return errors.New("nyi")
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
