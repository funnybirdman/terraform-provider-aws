package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
)

func TestAccAWSRDSClusterInstance_basic(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db:.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_maintenance_window"),
					resource.TestCheckResourceAttrSet(resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrSet(resourceName, "dbi_resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "engine", "aurora"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccAWSClusterInstanceConfigModified(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_isAlreadyBeingDeleted(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					// Get Database Instance into deleting state
					conn := testAccProvider.Meta().(*AWSClient).rdsconn
					input := &rds.DeleteDBInstanceInput{
						DBInstanceIdentifier: aws.String(fmt.Sprintf("tf-cluster-instance-%d", rInt)),
						SkipFinalSnapshot:    aws.Bool(true),
					}
					_, err := conn.DeleteDBInstance(input)
					if err != nil {
						t.Fatalf("error deleting Database Instance: %s", err)
					}
				},
				Config:  testAccAWSClusterInstanceConfig(rInt),
				Destroy: true,
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_az(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_az(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(resourceName, "availability_zone", regexp.MustCompile("^us-west-2[a-z]{1}$")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_namePrefix(t *testing.T) {
	var v rds.DBInstance
	rInt := acctest.RandInt()
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_namePrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestCheckResourceAttr(resourceName, "db_subnet_group_name", fmt.Sprintf("tf-test-%d", rInt)),
					resource.TestMatchResourceAttr(resourceName, "identifier", regexp.MustCompile("^tf-cluster-instance-")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_generatedName(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_generatedName(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccCheckAWSDBClusterInstanceAttributes(&v),
					resource.TestMatchResourceAttr(resourceName, "identifier", regexp.MustCompile("^tf-")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_kmsKey(t *testing.T) {
	var v rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.foo"
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigKmsKey(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/5350
func TestAccAWSRDSClusterInstance_disappears(t *testing.T) {
	var v rds.DBInstance
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig(acctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &v),
					testAccAWSClusterInstanceDisappears(&v),
				),
				// A non-empty plan is what we want. A crash is what we don't want. :)
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PubliclyAccessible(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterInstanceConfig_PubliclyAccessible(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccAWSRDSClusterInstanceConfig_PubliclyAccessible(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_CopyTagsToSnapshot(t *testing.T) {
	var dbInstance rds.DBInstance
	rNameSuffix := acctest.RandInt()
	resourceName := "aws_rds_cluster_instance.cluster_instances"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfig_CopyTagsToSnapshot(rNameSuffix, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccAWSClusterInstanceConfig_CopyTagsToSnapshot(rNameSuffix, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_snapshot", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSDBClusterInstanceAttributes(v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.Engine != "aurora" && *v.Engine != "aurora-postgresql" && *v.Engine != "aurora-mysql" {
			return fmt.Errorf("bad engine, expected \"aurora\", \"aurora-mysql\" or \"aurora-postgresql\": %#v", *v.Engine)
		}

		if !strings.HasPrefix(*v.DBClusterIdentifier, "tf-aurora-cluster") {
			return fmt.Errorf("Bad Cluster Identifier prefix:\nexpected: %s\ngot: %s", "tf-aurora-cluster", *v.DBClusterIdentifier)
		}

		return nil
	}
}

func testAccAWSClusterInstanceDisappears(v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).rdsconn
		opts := &rds.DeleteDBInstanceInput{
			DBInstanceIdentifier: v.DBInstanceIdentifier,
		}
		if _, err := conn.DeleteDBInstance(opts); err != nil {
			return err
		}
		return resource.Retry(40*time.Minute, func() *resource.RetryError {
			opts := &rds.DescribeDBInstancesInput{
				DBInstanceIdentifier: v.DBInstanceIdentifier,
			}
			_, err := conn.DescribeDBInstances(opts)
			if err != nil {
				dbinstanceerr, ok := err.(awserr.Error)
				if ok && dbinstanceerr.Code() == rds.ErrCodeDBInstanceNotFoundFault {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving DB Instances: %s", err))
			}
			return resource.RetryableError(fmt.Errorf(
				"Waiting for instance to be deleted: %v", v.DBInstanceIdentifier))
		})
	}
}

func testAccCheckAWSClusterInstanceExists(n string, v *rds.DBInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Instance ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn
		resp, err := conn.DescribeDBInstances(&rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, d := range resp.DBInstances {
			if *d.DBInstanceIdentifier == rs.Primary.ID {
				*v = *d
				return nil
			}
		}

		return fmt.Errorf("DB Cluster (%s) not found", rs.Primary.ID)
	}
}

func TestAccAWSRDSClusterInstance_MonitoringInterval(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "60"),
				),
			},
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "30"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_MonitoringRoleArn_EnabledToDisabled(t *testing.T) {
	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccAWSClusterInstanceConfigMonitoringInterval(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "monitoring_interval", "0"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_MonitoringRoleArn_EnabledToRemoved(t *testing.T) {
	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArnRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_MonitoringRoleArn_RemovedToEnabled(t *testing.T) {
	var dbInstance rds.DBInstance
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArnRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config: testAccAWSClusterInstanceConfigMonitoringRoleArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring_role_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsEnabled_AuroraMysql1(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsEnabled_AuroraMysql2(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsEnabled_AuroraPostgresql(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraPostgresql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraMysql1(t *testing.T) {
	var dbInstance rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraMysql1_DefaultKeyToCustomKey(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config:      testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql1(rName),
				ExpectError: regexp.MustCompile(`InvalidParameterCombination: You cannot change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraMysql2(t *testing.T) {
	var dbInstance rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraMysql2_DefaultKeyToCustomKey(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config:      testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql2(rName),
				ExpectError: regexp.MustCompile(`InvalidParameterCombination: You cannot change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraPostgresql(t *testing.T) {
	var dbInstance rds.DBInstance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraPostgresql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "performance_insights_kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_PerformanceInsightsKmsKeyId_AuroraPostgresql_DefaultKeyToCustomKey(t *testing.T) {
	var dbInstance rds.DBInstance
	resourceName := "aws_rds_cluster_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraPostgresql(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "performance_insights_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
			{
				Config:      testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraPostgresql(rName),
				ExpectError: regexp.MustCompile(`InvalidParameterCombination: You cannot change your Performance Insights KMS key`),
			},
		},
	})
}

func TestAccAWSRDSClusterInstance_CACertificateIdentifier(t *testing.T) {
	var dbInstance rds.DBInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterInstanceConfig_CACertificateIdentifier(rName, "rds-ca-2019"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSClusterInstanceExists(resourceName, &dbInstance),
					resource.TestCheckResourceAttr(resourceName, "ca_cert_identifier", "rds-ca-2019"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
					"identifier_prefix",
				},
			},
		},
	})
}

// Add some random to the name, to avoid collision
func testAccAWSClusterInstanceConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier  = "tf-aurora-cluster-test-%d"
  availability_zones  = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%d"
  cluster_identifier      = aws_rds_cluster.default.id
  instance_class          = "db.t2.small"
  db_parameter_group_name = aws_db_parameter_group.bar.name
  promotion_tier          = "3"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n)
}

func testAccAWSClusterInstanceConfigModified(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier  = "tf-aurora-cluster-test-%d"
  availability_zones  = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier                 = "tf-cluster-instance-%d"
  cluster_identifier         = aws_rds_cluster.default.id
  instance_class             = "db.t2.small"
  db_parameter_group_name    = aws_db_parameter_group.bar.name
  auto_minor_version_upgrade = false
  promotion_tier             = "3"
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n)
}

func testAccAWSClusterInstanceConfig_az(n int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_rds_cluster" "default" {
  cluster_identifier  = "tf-aurora-cluster-test-%d"
  availability_zones  = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%d"
  cluster_identifier      = aws_rds_cluster.default.id
  instance_class          = "db.t2.small"
  db_parameter_group_name = aws_db_parameter_group.bar.name
  promotion_tier          = "3"
  availability_zone       = data.aws_availability_zones.available.names[0]
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n)
}

func testAccAWSClusterInstanceConfig_namePrefix(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  identifier_prefix  = "tf-cluster-instance-"
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = "db.t2.small"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = "tf-aurora-cluster-%d"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-rds-cluster-instance-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-rds-cluster-instance-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-rds-cluster-instance-name-prefix-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "tf-test-%d"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, n, n)
}

func testAccAWSClusterInstanceConfig_generatedName(n int) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = "db.t2.small"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier   = "tf-aurora-cluster-%d"
  master_username      = "root"
  master_password      = "password"
  db_subnet_group_name = aws_db_subnet_group.test.name
  skip_final_snapshot  = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-rds-cluster-instance-generated-name"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-rds-cluster-instance-generated-name-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-rds-cluster-instance-generated-name-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "tf-test-%d"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}
`, n, n)
}

func testAccAWSClusterInstanceConfigKmsKey(n int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %d"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_rds_cluster" "default" {
  cluster_identifier  = "tf-aurora-cluster-test-%d"
  availability_zones  = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  storage_encrypted   = true
  kms_key_id          = aws_kms_key.foo.arn
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier              = "tf-cluster-instance-%d"
  cluster_identifier      = aws_rds_cluster.default.id
  instance_class          = "db.t2.small"
  db_parameter_group_name = aws_db_parameter_group.bar.name
}

resource "aws_db_parameter_group" "bar" {
  name   = "tfcluster-test-group-%d"
  family = "aurora5.6"

  parameter {
    name         = "back_log"
    value        = "32767"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, n, n, n, n)
}

func testAccAWSClusterInstanceConfigMonitoringInterval(rName string, monitoringInterval int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  instance_class      = "db.t2.small"
  monitoring_interval = %[2]d
  monitoring_role_arn = aws_iam_role.test.arn
}
`, rName, monitoringInterval)
}

func testAccAWSClusterInstanceConfigMonitoringRoleArnRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier = aws_rds_cluster.test.id
  identifier         = %[1]q
  instance_class     = "db.t2.small"
}
`, rName)
}

func testAccAWSClusterInstanceConfigMonitoringRoleArn(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "monitoring.rds.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %[1]q
  instance_class      = "db.t2.small"
  monitoring_interval = 5
  monitoring_role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql1(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = "aurora"
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  identifier                   = %[1]q
  instance_class               = "db.r4.large"
  performance_insights_enabled = true
}
`, rName)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraMysql2(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = "aurora-mysql"
  engine_version      = "5.7.mysql_aurora.2.04.2"
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  engine_version               = aws_rds_cluster.test.engine_version
  identifier                   = %[1]q
  instance_class               = "db.r4.large"
  performance_insights_enabled = true
}
`, rName)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsEnabledAuroraPostgresql(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = "aurora-postgresql"
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier           = aws_rds_cluster.test.id
  engine                       = aws_rds_cluster.test.engine
  identifier                   = %[1]q
  instance_class               = "db.r4.large"
  performance_insights_enabled = true
}
`, rName)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql1(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = "aurora"
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  identifier                      = %[1]q
  instance_class                  = "db.r4.large"
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraMysql2(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = "aurora-mysql"
  engine_version      = "5.7.mysql_aurora.2.04.2"
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  engine_version                  = aws_rds_cluster.test.engine_version
  identifier                      = %[1]q
  instance_class                  = "db.r4.large"
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName)
}

func testAccAWSClusterInstanceConfigPerformanceInsightsKmsKeyIdAuroraPostgresql(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  database_name       = "mydb"
  engine              = "aurora-postgresql"
  master_password     = "mustbeeightcharacters"
  master_username     = "foo"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier              = aws_rds_cluster.test.id
  engine                          = aws_rds_cluster.test.engine
  identifier                      = %[1]q
  instance_class                  = "db.r4.large"
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.test.arn
}
`, rName)
}

func testAccAWSRDSClusterInstanceConfig_PubliclyAccessible(rName string, publiclyAccessible bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately   = true
  cluster_identifier  = aws_rds_cluster.test.id
  identifier          = %q
  instance_class      = "db.t2.small"
  publicly_accessible = %t
}
`, rName, rName, publiclyAccessible)
}

func testAccAWSClusterInstanceConfig_CopyTagsToSnapshot(n int, f bool) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "default" {
  cluster_identifier  = "tf-aurora-cluster-test-%d"
  availability_zones  = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  identifier            = "tf-cluster-instance-%d"
  cluster_identifier    = aws_rds_cluster.default.id
  instance_class        = "db.t2.small"
  promotion_tier        = "3"
  copy_tags_to_snapshot = %t
}
`, n, n, f)
}

func testAccAWSRDSClusterInstanceConfig_CACertificateIdentifier(rName string, caCertificateIdentifier string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %q
  master_username     = "foo"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.test.id
  identifier         = %q
  instance_class     = "db.t2.small"
  ca_cert_identifier = %q
}
`, rName, rName, caCertificateIdentifier)
}
