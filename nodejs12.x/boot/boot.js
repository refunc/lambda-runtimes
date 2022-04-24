#!/usr/bin/env node
/** Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved. */

const lambda = require("@javascriptbach/aws-lambda-ric/lib/index");
const lambdaCommon = require("@javascriptbach/aws-lambda-ric/lib/Common/index");
const lambdaUserFunction = require("@javascriptbach/aws-lambda-ric/lib/utils/UserFunction");

const appRoot = process.cwd();
var handler = process.env._HANDLER;

if (!handler) {
    handler = process.env.AWS_LAMBDA_FUNCTION_HANDLER;
}

handlerFunc = appRoot

try {
    handlerFunc = lambdaCommon.isHandlerFunction(appRoot) ? appRoot : lambdaUserFunction.load(appRoot, handler)
} catch (e) {
    handlerFunc = async (event, context) => {
        throw e
    }
}

lambda.run(handlerFunc, handler);