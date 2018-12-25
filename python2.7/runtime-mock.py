from __future__ import print_function
import sys
import os
import random
import uuid
import time
import json
import resource
import datetime

orig_stdout = sys.stdout
orig_stderr = sys.stderr

from lambda_runtime_client import LambdaRuntimeClient


def eprint(*args, **kwargs):
    print(*args, file=orig_stderr, **kwargs)


def _random_account_id():
    return random.randint(100000000000, 999999999999)


def _random_invoke_id():
    return str(uuid.uuid4())


def _arn(region, account_id, fct_name):
    return "arn:aws:lambda:%s:%s:function:%s" % (region, account_id, fct_name)


_GLOBAL_HANDLER = (
    sys.argv[1]
    if len(sys.argv) > 1
    else os.environ.get(
        "AWS_LAMBDA_FUNCTION_HANDLER",
        os.environ.get("_HANDLER", "lambda_function.lambda_handler"),
    )
)

_GLOBAL_FCT_NAME = os.environ.get("AWS_LAMBDA_FUNCTION_NAME", "test")
_GLOBAL_VERSION = os.environ.get("AWS_LAMBDA_FUNCTION_VERSION", "$LATEST")
_GLOBAL_MEM_SIZE = os.environ.get("AWS_LAMBDA_FUNCTION_MEMORY_SIZE", "1536")
_GLOBAL_TIMEOUT = int(os.environ.get("AWS_LAMBDA_FUNCTION_TIMEOUT", "300"))

_GLOBAL_ACCESS_KEY_ID = os.environ.get("AWS_ACCESS_KEY_ID", "SOME_ACCESS_KEY_ID")
_GLOBAL_SECRET_ACCESS_KEY = os.environ.get(
    "AWS_SECRET_ACCESS_KEY", "SOME_SECRET_ACCESS_KEY"
)
_GLOBAL_SESSION_TOKEN = os.environ.get("AWS_SESSION_TOKEN", None)

_GLOBAL_INVOKEID = _random_invoke_id()
_GLOBAL_MODE = "event"  # Either 'http' or 'event'
_GLOBAL_SUPRESS_INIT = True  # Forces calling _get_handlers_delayed()
_GLOBAL_DATA_SOCK = -1
_GLOBAL_CREDENTIALS = {
    "key": _GLOBAL_ACCESS_KEY_ID,
    "secret": _GLOBAL_SECRET_ACCESS_KEY,
    "session": _GLOBAL_SESSION_TOKEN,
}
_GLOBAL_XRAY_TRACE_ID = os.environ.get("_X_AMZN_TRACE_ID", None)

_GLOBAL_START_TIME = None

lambda_runtime_api_addr = os.environ["AWS_LAMBDA_RUNTIME_API"]
del os.environ["AWS_LAMBDA_RUNTIME_API"]
lambda_runtime_client = LambdaRuntimeClient(lambda_runtime_api_addr)


def report_user_init_start():
    return


def report_user_init_end():
    return


def report_user_invoke_start():
    return


def report_user_invoke_end():
    return


def receive_start():
    sys.stdout = orig_stderr
    sys.stderr = orig_stderr
    return (
        _GLOBAL_INVOKEID,
        _GLOBAL_MODE,
        _GLOBAL_HANDLER,
        _GLOBAL_SUPRESS_INIT,
        _GLOBAL_CREDENTIALS,
    )


def report_running(invokeid):
    return


def receive_invoke():
    global _GLOBAL_START_TIME
    global _GLOBAL_INVOKEID

    event_request = lambda_runtime_client.wait_next_invocation()

    _GLOBAL_INVOKEID = event_request.invoke_id

    eprint("START RequestId: %s Version: %s" % (_GLOBAL_INVOKEID, _GLOBAL_VERSION))

    _GLOBAL_START_TIME = time.time()

    if event_request.cognito_identity:
        cognito_identity = json.loads(event_request.cognito_identity)
        cognitoidentityid, cognitopoolid = (
            cognito_identity["cognitoidentityid"],
            cognito_identity["cognitopoolid"],
        )
    else:
        cognitoidentityid, cognitopoolid = None, None

    return (
        event_request.invoke_id,
        _GLOBAL_DATA_SOCK,
        _GLOBAL_CREDENTIALS,
        event_request.event_body,
        {
            "clientcontext": event_request.client_context,
            "cognitoidentityid": cognitoidentityid,
            "cognitopoolid": cognitopoolid,
        },
        event_request.invoked_function_arn,
        _GLOBAL_XRAY_TRACE_ID,
    )


def report_fault(invokeid, msg, except_value, trace):
    if msg and except_value:
        eprint("%s: %s" % (msg, except_value))
    if trace:
        eprint("%s" % trace)

    return


def report_done(invokeid, errortype, result, is_fatal):
    if not result:
        return

    eprint("END RequestId: %s" % invokeid)

    duration = int((time.time() - _GLOBAL_START_TIME) * 1000)
    billed_duration = min(100 * int((duration / 100) + 1), _GLOBAL_TIMEOUT * 1000)
    max_mem = int(resource.getrusage(resource.RUSAGE_SELF).ru_maxrss / 1024)

    eprint(
        "REPORT RequestId: %s Duration: %s ms Billed Duration: %s ms Memory Size: %s MB Max Memory Used: %s MB"
        % (invokeid, duration, billed_duration, _GLOBAL_MEM_SIZE, max_mem)
    )
    if errortype:
        lambda_runtime_client.post_invocation_error(invokeid, result)
    else:
        lambda_runtime_client.post_invocation_result(invokeid, result)


def report_xray_exception(xray_json):
    return


def log_bytes(msg, fileno):
    eprint(msg)
    return


def log_sb(msg):
    return


def get_remaining_time():
    return (_GLOBAL_TIMEOUT * 1000) - int((time.time() - _GLOBAL_START_TIME) * 1000)


def send_console_message(msg, byte_length):
    eprint(msg)
    return


def make_error(errorMessage, errorType, stackTrace):  # stackTrace is an array
    result = {}
    if errorMessage:
        result["errorMessage"] = errorMessage
    if errorType:
        result["errorType"] = errorType
    if stackTrace:
        result["stackTrace"] = stackTrace
    return result
