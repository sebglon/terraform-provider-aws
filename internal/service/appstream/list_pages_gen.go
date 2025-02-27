// Code generated by "internal/generate/listpages/main.go -ListOps=DescribeDirectoryConfigs,DescribeFleets,DescribeImageBuilders,DescribeStacks,DescribeUsers,ListAssociatedStacks -ContextOnly"; DO NOT EDIT.

package appstream

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/aws/aws-sdk-go/service/appstream/appstreamiface"
)

func describeDirectoryConfigsPages(ctx context.Context, conn appstreamiface.AppStreamAPI, input *appstream.DescribeDirectoryConfigsInput, fn func(*appstream.DescribeDirectoryConfigsOutput, bool) bool) error {
	for {
		output, err := conn.DescribeDirectoryConfigsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
func describeFleetsPages(ctx context.Context, conn appstreamiface.AppStreamAPI, input *appstream.DescribeFleetsInput, fn func(*appstream.DescribeFleetsOutput, bool) bool) error {
	for {
		output, err := conn.DescribeFleetsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
func describeImageBuildersPages(ctx context.Context, conn appstreamiface.AppStreamAPI, input *appstream.DescribeImageBuildersInput, fn func(*appstream.DescribeImageBuildersOutput, bool) bool) error {
	for {
		output, err := conn.DescribeImageBuildersWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
func describeStacksPages(ctx context.Context, conn appstreamiface.AppStreamAPI, input *appstream.DescribeStacksInput, fn func(*appstream.DescribeStacksOutput, bool) bool) error {
	for {
		output, err := conn.DescribeStacksWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
func describeUsersPages(ctx context.Context, conn appstreamiface.AppStreamAPI, input *appstream.DescribeUsersInput, fn func(*appstream.DescribeUsersOutput, bool) bool) error {
	for {
		output, err := conn.DescribeUsersWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
func listAssociatedStacksPages(ctx context.Context, conn appstreamiface.AppStreamAPI, input *appstream.ListAssociatedStacksInput, fn func(*appstream.ListAssociatedStacksOutput, bool) bool) error {
	for {
		output, err := conn.ListAssociatedStacksWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
