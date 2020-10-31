package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalytics/finder"
)

func init() {
	resource.AddTestSweepers("aws_kinesis_analytics_application", &resource.Sweeper{
		Name: "aws_kinesis_analytics_application",
		F:    testSweepKinesisAnalyticsApplications,
	})
}

func testSweepKinesisAnalyticsApplications(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).kinesisanalyticsconn
	input := &kinesisanalytics.ListApplicationsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListApplications(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Kinesis Analytics Application sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Kinesis Analytics Applications: %w", err))
			return sweeperErrs
		}

		var name string
		for _, applicationSummary := range output.ApplicationSummaries {
			arn := aws.StringValue(applicationSummary.ApplicationARN)
			name = aws.StringValue(applicationSummary.ApplicationName)

			application, err := finder.ApplicationByName(conn, name)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading Kinesis Analytics Application (%s): %w", arn, err)
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			r := resourceAwsKinesisAnalyticsApplication()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("create_timestamp", aws.TimeValue(application.CreateTimestamp).Format(time.RFC3339))
			d.Set("name", name)
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		if !aws.BoolValue(output.HasMoreApplications) {
			break
		}
		input.ExclusiveStartApplicationName = aws.String(name)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSKinesisAnalyticsApplication_basic(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_disappears(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsKinesisAnalyticsApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_Tags(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_Code_Update(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigCode(rName, "SELECT 1;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", "SELECT 1;\n"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigCode(rName, "SELECT 2;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", "SELECT 2;\n"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_update(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplication_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplication_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "version", "2"),
					resource.TestCheckResourceAttr(resName, "code", "testCode2\n"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_addCloudwatchLoggingOptions(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()
	firstStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_basic(rInt)
	thirdStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_cloudwatchLoggingOptions(rInt, "testStream")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: firstStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "version", "1"),
				),
			},
			{
				Config: thirdStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "version", "2"),
					resource.TestCheckResourceAttr(resName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resName, "cloudwatch_logging_options.0.log_stream_arn", "aws_cloudwatch_log_stream.test", "arn"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_updateCloudwatchLoggingOptions(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()
	firstStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_cloudwatchLoggingOptions(rInt, "testStream")
	secondStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_cloudwatchLoggingOptions(rInt, "testStream2")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: firstStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "version", "1"),
					resource.TestCheckResourceAttr(resName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resName, "cloudwatch_logging_options.0.log_stream_arn", "aws_cloudwatch_log_stream.test", "arn"),
				),
			},
			{
				Config: secondStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "version", "2"),
					resource.TestCheckResourceAttr(resName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resName, "cloudwatch_logging_options.0.log_stream_arn", "aws_cloudwatch_log_stream.test", "arn"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_inputsKinesisFirehose(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_inputsKinesisFirehose(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.kinesis_firehose.#", "1"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_inputsKinesisStream(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_inputsKinesisStream(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "version", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.name_prefix", "test_prefix"),
					resource.TestCheckResourceAttr(resName, "inputs.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_inputsAdd(t *testing.T) {
	var before, after kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()
	firstStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_basic(rInt)
	secondStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_inputsKinesisStream(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: firstStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "version", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.#", "0"),
				),
			},
			{
				Config: secondStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &after),
					resource.TestCheckResourceAttr(resName, "version", "2"),
					resource.TestCheckResourceAttr(resName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.name_prefix", "test_prefix"),
					resource.TestCheckResourceAttr(resName, "inputs.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_inputsUpdateKinesisStream(t *testing.T) {
	var before, after kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()
	firstStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_inputsKinesisStream(rInt)
	secondStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_inputsUpdateKinesisStream(rInt, "testStream")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: firstStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "version", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.name_prefix", "test_prefix"),
					resource.TestCheckResourceAttr(resName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
				),
			},
			{
				Config: secondStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &after),
					resource.TestCheckResourceAttr(resName, "version", "2"),
					resource.TestCheckResourceAttr(resName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resName, "inputs.0.name_prefix", "test_prefix2"),
					resource.TestCheckResourceAttrPair(resName, "inputs.0.kinesis_stream.0.resource_arn", "aws_kinesis_stream.test", "arn"),
					resource.TestCheckResourceAttr(resName, "inputs.0.parallelism.0.count", "2"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.0.record_columns.0.name", "test2"),
					resource.TestCheckResourceAttr(resName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_outputsKinesisStream(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()
	firstStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_outputsKinesisStream(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: firstStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "version", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.name", "test_name"),
					resource.TestCheckResourceAttr(resName, "outputs.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.schema.0.record_format_type", "JSON"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_outputsMultiple(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()
	step := testAccKinesisAnalyticsApplication_prereq(rInt1) + testAccKinesisAnalyticsApplication_outputsMultiple(rInt1, rInt2)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: step,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "outputs.#", "2"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_outputsAdd(t *testing.T) {
	var before, after kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()
	firstStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_basic(rInt)
	secondStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_outputsKinesisStream(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: firstStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "version", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.#", "0"),
				),
			},
			{
				Config: secondStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &after),
					resource.TestCheckResourceAttr(resName, "version", "2"),
					resource.TestCheckResourceAttr(resName, "outputs.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.name", "test_name"),
					resource.TestCheckResourceAttr(resName, "outputs.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.schema.#", "1"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_outputsUpdateKinesisStream(t *testing.T) {
	var before, after kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()
	firstStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_outputsKinesisStream(rInt)
	secondStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_outputsUpdateKinesisStream(rInt, "testStream")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: firstStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "version", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.name", "test_name"),
					resource.TestCheckResourceAttr(resName, "outputs.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.schema.0.record_format_type", "JSON"),
				),
			},
			{
				Config: secondStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &after),
					resource.TestCheckResourceAttr(resName, "version", "2"),
					resource.TestCheckResourceAttr(resName, "outputs.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.name", "test_name2"),
					resource.TestCheckResourceAttr(resName, "outputs.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttrPair(resName, "outputs.0.kinesis_stream.0.resource_arn", "aws_kinesis_stream.test", "arn"),
					resource.TestCheckResourceAttr(resName, "outputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resName, "outputs.0.schema.0.record_format_type", "CSV"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_Outputs_Lambda_Add(t *testing.T) {
	var application1, application2 kinesisanalytics.ApplicationDetail
	iamRoleResourceName := "aws_iam_role.kinesis_analytics_application"
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigOutputsLambda(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &application2),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.lambda.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "outputs.0.lambda.0.resource_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "outputs.0.lambda.0.role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_Outputs_Lambda_Create(t *testing.T) {
	var application1 kinesisanalytics.ApplicationDetail
	iamRoleResourceName := "aws_iam_role.kinesis_analytics_application"
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigOutputsLambda(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.lambda.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "outputs.0.lambda.0.resource_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "outputs.0.lambda.0.role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_referenceDataSource(t *testing.T) {
	var application kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()
	firstStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_referenceDataSource(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: firstStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &application),
					resource.TestCheckResourceAttr(resName, "version", "2"),
					resource.TestCheckResourceAttr(resName, "reference_data_sources.#", "1"),
					resource.TestCheckResourceAttr(resName, "reference_data_sources.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resName, "reference_data_sources.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resName, "reference_data_sources.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_referenceDataSourceUpdate(t *testing.T) {
	var before, after kinesisanalytics.ApplicationDetail
	resName := "aws_kinesis_analytics_application.test"
	rInt := acctest.RandInt()
	firstStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_referenceDataSource(rInt)
	secondStep := testAccKinesisAnalyticsApplication_prereq(rInt) + testAccKinesisAnalyticsApplication_referenceDataSourceUpdate(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: firstStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &before),
					resource.TestCheckResourceAttr(resName, "version", "2"),
					resource.TestCheckResourceAttr(resName, "reference_data_sources.#", "1"),
				),
			},
			{
				Config: secondStep,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resName, &after),
					resource.TestCheckResourceAttr(resName, "version", "3"),
					resource.TestCheckResourceAttr(resName, "reference_data_sources.#", "1"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckKinesisAnalyticsApplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_analytics_application" {
			continue
		}

		_, err := finder.ApplicationByName(conn, rs.Primary.Attributes["name"])
		if isAWSErr(err, kinesisanalytics.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Kinesis Analytics Application %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckKinesisAnalyticsApplicationExists(n string, v *kinesisanalytics.ApplicationDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Analytics Application ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn

		application, err := finder.ApplicationByName(conn, rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		*v = *application

		return nil
	}
}

func testAccPreCheckAWSKinesisAnalytics(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn

	input := &kinesisanalytics.ListApplicationsInput{}

	_, err := conn.ListApplications(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccKinesisAnalyticsApplicationConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q
}
`, rName)
}

func testAccKinesisAnalyticsApplicationConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccKinesisAnalyticsApplicationConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccKinesisAnalyticsApplicationConfigCode(rName, code string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name        = %[1]q
  description = "test"
  code        = %[2]q
}
`, rName, code)
}

func testAccKinesisAnalyticsApplication_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"
}
`, rInt)
}

func testAccKinesisAnalyticsApplication_update(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode2\n"
}
`, rInt)
}

func testAccKinesisAnalyticsApplication_cloudwatchLoggingOptions(rInt int, streamName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "testAcc-%d"
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = "testAcc-%s-%d"
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.arn
    role_arn       = aws_iam_role.test.arn
  }
}
`, rInt, streamName, rInt, rInt)
}

func testAccKinesisAnalyticsApplication_inputsKinesisFirehose(rInt int) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "trust_firehose" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["firehose.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "firehose" {
  name               = "testAcc-firehose-%d"
  assume_role_policy = data.aws_iam_policy_document.trust_firehose.json
}

data "aws_iam_policy_document" "trust_lambda" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "testAcc-lambda-%d"
  assume_role_policy = data.aws_iam_policy_document.trust_lambda.json
}

resource "aws_s3_bucket" "test" {
  bucket = "testacc-%d"
  acl    = "private"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "testAcc-%d"
  handler       = "exports.example"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = "testAcc-%d"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"

  inputs {
    name_prefix = "test_prefix"

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test.arn
    }

    parallelism {
      count = 1
    }

    schema {
      record_columns {
        mapping  = "$.test"
        name     = "test"
        sql_type = "VARCHAR(8)"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "\n"
          }
        }
      }
    }
  }
}
`, rInt, rInt, rInt, rInt, rInt, rInt)
}

func testAccKinesisAnalyticsApplication_inputsKinesisStream(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = "testAcc-%d"
  shard_count = 1
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"

  inputs {
    name_prefix = "test_prefix"

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test.arn
    }

    parallelism {
      count = 1
    }

    schema {
      record_columns {
        mapping  = "$.test"
        name     = "test"
        sql_type = "VARCHAR(8)"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          json {
            record_row_path = "$"
          }
        }
      }
    }
  }
}
`, rInt, rInt)
}

func testAccKinesisAnalyticsApplication_inputsUpdateKinesisStream(rInt int, streamName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = "testAcc-%s-%d"
  shard_count = 1
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"

  inputs {
    name_prefix = "test_prefix2"

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test.arn
    }

    parallelism {
      count = 2
    }

    schema {
      record_columns {
        mapping  = "$.test2"
        name     = "test2"
        sql_type = "VARCHAR(8)"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "\n"
          }
        }
      }
    }
  }
}
`, streamName, rInt, rInt)
}

func testAccKinesisAnalyticsApplication_outputsKinesisStream(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = "testAcc-%d"
  shard_count = 1
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"

  outputs {
    name = "test_name"

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test.arn
    }

    schema {
      record_format_type = "JSON"
    }
  }
}
`, rInt, rInt)
}

func testAccKinesisAnalyticsApplication_outputsMultiple(rInt1, rInt2 int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test1" {
  name        = "testAcc-%d"
  shard_count = 1
}

resource "aws_kinesis_stream" "test2" {
  name        = "testAcc-%d"
  shard_count = 1
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"

  outputs {
    name = "test_name1"

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test1.arn
      role_arn     = aws_iam_role.test.arn
    }

    schema {
      record_format_type = "JSON"
    }
  }

  outputs {
    name = "test_name2"

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test2.arn
      role_arn     = aws_iam_role.test.arn
    }

    schema {
      record_format_type = "JSON"
    }
  }
}
`, rInt1, rInt2, rInt1)
}

func testAccKinesisAnalyticsApplicationConfigOutputsLambda(rInt int) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "kinesisanalytics_assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["kinesisanalytics.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "lambda_assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "kinesis_analytics_application" {
  name               = "tf-acc-test-%d-kinesis"
  assume_role_policy = data.aws_iam_policy_document.kinesisanalytics_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "kinesis_analytics_application-AWSLambdaRole" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaRole"
  role       = aws_iam_role.kinesis_analytics_application.name
}

resource "aws_iam_role" "lambda_function" {
  name               = "tf-acc-test-%d-lambda"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "lambda_function-AWSLambdaBasicExecutionRole" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.lambda_function.name
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "tf-acc-test-%d"
  handler       = "exports.example"
  role          = aws_iam_role.lambda_function.arn
  runtime       = "nodejs12.x"
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"

  outputs {
    name = "test_name"

    lambda {
      resource_arn = aws_lambda_function.test.arn
      role_arn     = aws_iam_role.kinesis_analytics_application.arn
    }

    schema {
      record_format_type = "JSON"
    }
  }
}
`, rInt, rInt, rInt, rInt)
}

func testAccKinesisAnalyticsApplication_outputsUpdateKinesisStream(rInt int, streamName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = "testAcc-%s-%d"
  shard_count = 1
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"
  code = "testCode\n"

  outputs {
    name = "test_name2"

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test.arn
    }

    schema {
      record_format_type = "CSV"
    }
  }
}
`, streamName, rInt, rInt)
}

func testAccKinesisAnalyticsApplication_referenceDataSource(rInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "testacc-%d"
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"

  reference_data_sources {
    table_name = "test_table"

    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = "test_file_key"
      role_arn   = aws_iam_role.test.arn
    }

    schema {
      record_columns {
        mapping  = "$.test"
        name     = "test"
        sql_type = "VARCHAR(8)"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          json {
            record_row_path = "$"
          }
        }
      }
    }
  }
}
`, rInt, rInt)
}

func testAccKinesisAnalyticsApplication_referenceDataSourceUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "testacc2-%d"
}

resource "aws_kinesis_analytics_application" "test" {
  name = "testAcc-%d"

  reference_data_sources {
    table_name = "test_table2"

    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = "test_file_key"
      role_arn   = aws_iam_role.test.arn
    }

    schema {
      record_columns {
        mapping  = "$.test2"
        name     = "test2"
        sql_type = "VARCHAR(8)"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "\n"
          }
        }
      }
    }
  }
}
`, rInt, rInt)
}

// this is used to set up the IAM role
func testAccKinesisAnalyticsApplication_prereq(rInt int) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "trust" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["kinesisanalytics.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = "testAcc-%d"
  assume_role_policy = data.aws_iam_policy_document.trust.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["firehose:*"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "test" {
  name   = "testAcc-%d"
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}
`, rInt, rInt)
}
