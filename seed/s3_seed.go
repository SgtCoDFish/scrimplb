package seed

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

// S3Provider can retrieve a seed from an object in an S3 bucket
type S3Provider struct {
	Bucket           string
	AvailabilityZone string
}

type s3SeedConfig struct {
	Bucket           string  `json:"bucket"`
	Region           string  `json:"region"`
	AvailabilityZone *string `json:"availabilityZone"`
}

// NewS3Provider creates a new S3 seed provider from a file called
// "seed.json" in the given configDir
func NewS3Provider(configDir string) (*S3Provider, error) {
	contents, err := ioutil.ReadFile(path.Join(configDir, "seed.json"))

	if err != nil {
		return nil, errors.Wrap(err, "couldn't open seed.json in expected location")
	}

	var result s3SeedConfig

	err = json.Unmarshal(contents, &result)

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse seed.json for S3 seed provider")
	}

	var az string

	if result.AvailabilityZone == nil {
		az, err = deduceInstanceAZ()

		if err != nil {
			return nil, errors.Wrap(err, "failed to deduce instance AZ")
		}
	} else {
		az = *result.AvailabilityZone
	}

	return &S3Provider{
		Bucket:           result.Bucket,
		AvailabilityZone: az,
		Region:           result.Region,
	}, nil
}

// FetchSeed makes a call to S3 to retrieve the seed for the AZ this
// instance is located in.
func (s *S3Provider) FetchSeed() (Seed, error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(s.Region)},
	)
	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(s.AvailabilityZone),
		})

	if err != nil {
		return Seed{}, fmt.Errorf("Unable to download item %q, %v", item, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	return Seed{
		"test",
		123,
	}, nil
}

func deduceInstanceAZ() (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/placement/availability-zone")

	if err != nil {
		return "", nil
	}

	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(raw), nil
}
