package AWSAT001

import (
	"go/ast"
	"strings"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/testmatchresourceattr"
	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/asthelper"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for resource.TestMatchResourceAttr() calls against ARN attributes

The AWSAT001 analyzer reports when a resource.TestMatchResourceAttr() call references an Amazon
Resource Name (ARN) attribute. It is preferred to use resource.TestCheckResourceAttrPair() or one
one of the available Terraform AWS Provider ARN testing check functions instead building full ARN
regular expressions. These testing helper functions consider the value of the AWS Account ID,
Partition, and Region of the acceptance test runner.

The resource.TestCheckResourceAttrPair() call can be used when the Terraform state has the ARN
value already available, such as when the current resource is referencing an ARN attribute of
another resource.

Otherwise, available ARN testing check functions include:

- testAccCheckResourceAttrGlobalARN
- testAccCheckResourceAttrGlobalARNNoAccount
- testAccCheckResourceAttrRegionalARN
- testAccMatchResourceAttrGlobalARN
- testAccMatchResourceAttrRegionalARN
- testAccMatchResourceAttrRegionalARNNoAccount
`

const analyzerName = "AWSAT001"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		testmatchresourceattr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	callExprs := pass.ResultOf[testmatchresourceattr.Analyzer].([]*ast.CallExpr)
	commentIgnorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)

	for _, callExpr := range callExprs {
		if commentIgnorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		attributeName := asthelper.AstBasicLitStringValue(callExpr.Args[1])

		if attributeName == "" {
			continue
		}

		if !AttributeNameAppearsArnRelated(attributeName) {
			continue
		}

		pass.Reportf(callExpr.Pos(), "%s: prefer resource.TestCheckResourceAttrPair() or ARN check functions (e.g. testAccMatchResourceAttrRegionalARN)", analyzerName)
	}

	return nil, nil
}

func AttributeNameAppearsArnRelated(attributeName string) bool {
	if attributeName == "arn" {
		return true
	}

	if strings.HasSuffix(attributeName, "_arn") {
		return true
	}

	// Handle flatmap nested attribute
	if strings.HasSuffix(attributeName, ".arn") {
		return true
	}

	return false
}
