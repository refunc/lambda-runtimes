import os
import sys

# docker run --rm -v "$PWD":/var/task refunc/lambda:python3.9 lambda_function.lambda_handler


def lambda_handler(event, context):

    if "error" in event:
        # test for exception
        raise ValueError(event["error"])

    context.log("Hello!")
    context.log("Hmmm, does not add newlines in 3.10?")
    context.log("\n")

    print(sys.executable)
    print(sys.argv)
    print(os.getcwd())
    print(__file__)
    print(os.environ)
    print(context.__dict__)
    print("--------------------")
    return {
        "executable": str(sys.executable),
        "sys.argv": str(sys.argv),
        "os.getcwd": str(os.getcwd()),
        "__file__": str(__file__),
        "os.environ": str(os.environ),
        "context.__dict__": str(context.__dict__),
        "event": event,
    }
