package seed

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sgtcodfish/scrimplb/constants"
	"github.com/sgtcodfish/scrimplb/resolver"
)

// S3Provider can retrieve a seed from an object in an S3 bucket
type S3Provider struct {
	Bucket string
	Region string
	Key    string

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
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.Region)},
	)

	if err != nil {
		return Seeds{}, err
	}

	downloader := s3manager.NewDownloader(sess)

	buf := aws.NewWriteAtBuffer([]byte{})

	numBytes, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(s.Key),
		})

	if err != nil {
		return Seeds{}, errors.Wrap(err, "unable to download seed")
	}

	log.Printf("downloaded %d bytes from S3 for seed\n", numBytes)

	var seeds Seeds

	err = json.Unmarshal(buf.Bytes(), &seeds)

	if err != nil {
		return Seeds{}, errors.Wrap(err, "unable to parse downloaded seed")
	}

	return seeds, nil
}

// PushSeed fetches remote data, updates with local node details and then
// publishes the details to an S3 bucket so that future nodes can join
func (s *S3Provider) PushSeed(resolver resolver.IPResolver, port string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.Region)},
	)

	if err != nil {
		return err
	}

	ip, err := resolver.ResolveIP()

	if err != nil {
		return err
	}

	out, err := json.Marshal(Seeds{
		Seeds: []Seed{{
			Address: ip,
			Port:    port,
		}},
	})

	if err != nil {
		return errors.Wrap(err, "couldn't marshal output for S3")
	}

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
		Body:   bytes.NewReader(out),
	})

	if err != nil {
		return errors.Wrap(err, "couldn't upload S3 content")
	}

	log.Println("successfully pushed seed to s3")

	return nil
}
