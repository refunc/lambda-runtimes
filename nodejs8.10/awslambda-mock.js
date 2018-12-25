var fs = require('fs')
var crypto = require('crypto')

const http = require('http')

const RUNTIME_PATH = '/2018-06-01/runtime'

const {
  AWS_LAMBDA_FUNCTION_NAME,
  AWS_LAMBDA_FUNCTION_VERSION,
  AWS_LAMBDA_FUNCTION_MEMORY_SIZE,
  AWS_LAMBDA_LOG_GROUP_NAME,
  AWS_LAMBDA_LOG_STREAM_NAME,
  AWS_LAMBDA_RUNTIME_API,
  AWS_ACCESS_KEY_ID,
  AWS_SECRET_ACCESS_KEY,
  AWS_SESSION_TOKEN,
  AWS_LAMBDA_FUNCTION_INVOKED_ARN,
} = process.env

const [HOST, PORT] = AWS_LAMBDA_RUNTIME_API.split(':')

var HANDLER = process.argv[2] || process.env.AWS_LAMBDA_FUNCTION_HANDLER || process.env._HANDLER || 'index.handler'
var FN_NAME = AWS_LAMBDA_FUNCTION_NAME || 'test'
var TIMEOUT = process.env.AWS_LAMBDA_FUNCTION_TIMEOUT || '300'
var REGION = process.env.AWS_REGION || process.env.AWS_DEFAULT_REGION || 'us-east-1'
var ACCOUNT_ID = process.env.AWS_ACCOUNT_ID || randomAccountId()
var INVOKED_ARN = AWS_LAMBDA_FUNCTION_INVOKED_ARN || arn(REGION, ACCOUNT_ID, FN_NAME)

function consoleLog(str) {
  process.stderr.write(formatConsole(str))
}

function systemLog(str) {
  process.stderr.write(formatSystem(str) + '\n')
}

function systemErr(str) {
  process.stderr.write(formatErr(str) + '\n')
}

// Don't think this can be done in the Docker image
process.umask(2)

var InitOptions = {
  initInvokeId: uuid(),
  invokeId: uuid(),
  handler: HANDLER,
  suppressInit: true,
  credentials: {
    key: AWS_ACCESS_KEY_ID,
    secret: AWS_SECRET_ACCESS_KEY,
    session: AWS_SESSION_TOKEN,
  },
  eventBody: "",
  contextObjects: {},
  invokedFunctionArn: INVOKED_ARN,
}

// Some weird spelling error in the source?
InitOptions.invokeid = InitOptions.invokeId

var invoked = false
var errored = false
var start = null

module.exports = {
  initRuntime: function() { return InitOptions },
  waitForInvoke: function(callback) {
    invoked = true
    nextInvocation(({event, context})=>{
      systemLog('START RequestId: ' + context.awsRequestId + ' Version: ' + AWS_LAMBDA_FUNCTION_VERSION)
      start = process.hrtime()
      callback({
        invokeId: context.awsRequestId,
        invokeid: context.awsRequestId,
        credentials: {
          key: AWS_ACCESS_KEY_ID,
          secret: AWS_SECRET_ACCESS_KEY,
          session: AWS_SESSION_TOKEN,
        },
        eventBody: event,
        contextObjects: {
          clientContext: context.clientContext,
          cognitoIdentityId: context.identity ? context.identity.cognitoIdentityId : undefined,
          cognitoPoolId: context.identity ? context.identity.cognitoPoolId : undefined,
        },
        invokedFunctionArn: context.invokedFunctionArn,
      })
    })
  },
  reportRunning: function(invokeId) {}, // eslint-disable-line no-unused-vars
  reportDone: function(awsRequestId, errType, resultStr) {
    if (!invoked) return
    var diffMs = hrTimeMs(process.hrtime(start))
    var billedMs = Math.min(100 * (Math.floor(diffMs / 100) + 1), TIMEOUT * 1000)
    systemLog('END RequestId: ' + awsRequestId)
    systemLog([
      'REPORT RequestId: ' + awsRequestId,
      'Duration: ' + diffMs.toFixed(2) + ' ms',
      'Billed Duration: ' + billedMs + ' ms',
      'Memory Size: ' + AWS_LAMBDA_FUNCTION_MEMORY_SIZE + ' MB',
      'Max Memory Used: ' + Math.round(process.memoryUsage().rss / (1024 * 1024)) + ' MB',
      '',
    ].join('\t'))

    if (errored || errType) {
      postError(resultStr, {awsRequestId})
    } else {
      invokeResponse(resultStr, {awsRequestId})
    }
  },
  reportFault: function(invokeId, msg, errName, errStack) {
    errored = true
    systemErr(msg + (errName ? ': ' + errName : ''))
    if (errStack) systemErr(errStack)
  },
  reportUserInitStart: function() {},
  reportUserInitEnd: function() {},
  reportUserInvokeStart: function() {},
  reportUserInvokeEnd: function() {},
  reportException: function() {},
  getRemainingTime: function() {
    return (TIMEOUT * 1000) - Math.floor(hrTimeMs(process.hrtime(start)))
  },
  sendConsoleLogs: consoleLog,
  maxLoggerErrorSize: 256 * 1024,
}

function formatConsole(str) {
  return str.replace(/^[0-9TZ:.-]+\t[0-9a-f-]+\t/, '\u001b[34m$&\u001b[0m')
}

function formatSystem(str) {
  return '\u001b[32m' + str + '\u001b[0m'
}

function formatErr(str) {
  return '\u001b[31m' + str + '\u001b[0m'
}

function hrTimeMs(hrtime) {
  return (hrtime[0] * 1e9 + hrtime[1]) / 1e6
}

// Approximates the look of a v1 UUID
function uuid() {
  return crypto.randomBytes(4).toString('hex') + '-' +
    crypto.randomBytes(2).toString('hex') + '-' +
    crypto.randomBytes(2).toString('hex').replace(/^./, '1') + '-' +
    crypto.randomBytes(2).toString('hex') + '-' +
    crypto.randomBytes(6).toString('hex')
}

function randomAccountId() {
  return String(0x100000000 * Math.random())
}

function arn(region, accountId, fnName) {
  return 'arn:aws:lambda:' + region + ':' + accountId.replace(/[^\d]/g, '') + ':function:' + fnName
}

function request(options) {
  options.host = HOST
  options.port = PORT

  return new Promise((resolve, reject) => {
    let req = http.request(options, res => {
      let bufs = []
      res.on('data', data => bufs.push(data))
      res.on('end', () => resolve({
        statusCode: res.statusCode,
        headers: res.headers,
        body: Buffer.concat(bufs).toString(),
      }))
      res.on('error', reject)
    })
    req.on('error', reject)
    req.end(options.body)
  })
}


function nextInvocation(callback) {
  request({ path: `${RUNTIME_PATH}/invocation/next` }).then((res) => {
    if (res.statusCode !== 200) {
      throw new Error(`Unexpected /invocation/next response: ${JSON.stringify(res)}`)
    }

    let traceId
    if (res.headers['lambda-runtime-trace-id']) {
      process.env._X_AMZN_TRACE_ID = res.headers['lambda-runtime-trace-id']
      traceId = res.headers['lambda-runtime-trace-id']
    } else {
      delete process.env._X_AMZN_TRACE_ID
    }

    const deadlineMs = +res.headers['lambda-runtime-deadline-ms']

    let context = {
      traceId,
      awsRequestId: res.headers['lambda-runtime-aws-request-id'],
      invokedFunctionArn: res.headers['lambda-runtime-invoked-function-arn'],
      logGroupName: AWS_LAMBDA_LOG_GROUP_NAME,
      logStreamName: AWS_LAMBDA_LOG_STREAM_NAME,
      functionName: AWS_LAMBDA_FUNCTION_NAME,
      functionVersion: AWS_LAMBDA_FUNCTION_VERSION,
      memoryLimitInMB: AWS_LAMBDA_FUNCTION_MEMORY_SIZE,
      getRemainingTimeInMillis: () => deadlineMs - Date.now(),
    }

    if (res.headers['lambda-runtime-client-context']) {
      context.clientContext = res.headers['lambda-runtime-client-context']
    }

    if (res.headers['lambda-runtime-cognito-identity']) {
      context.identity = JSON.parse(res.headers['lambda-runtime-cognito-identity'])
    }

    const event = res.body
    callback({ event, context })

  })
}

function invokeResponse(result, context) {
  request({
    method: 'POST',
    path: `${RUNTIME_PATH}/invocation/${context.awsRequestId}/response`,
    body: result,
  }).then((res)=>{
    if (res.statusCode !== 202) {
      throw new Error(`Unexpected /invocation/response response: ${JSON.stringify(res)}`)
    }
  })
}

function postError(lambdaErr, context) {
  request({
    method: 'POST',
    path: `${RUNTIME_PATH}/invocation/${context.awsRequestId}/error`,
    headers: {
      'Content-Type': 'application/json',
      'Lambda-Runtime-Function-Error-Type': lambdaErr.errorType || 'Unexpected',
    },
    body: JSON.stringify(lambdaErr),
  }).then((res) => {
    if (res.statusCode !== 202) {
      throw new Error(`Unexpected ${path} response: ${JSON.stringify(res)}`)
    }
  })
}

process.on('unhandledRejection', function(err) {
  console.error(err.stack);
  process.exit(1);
});