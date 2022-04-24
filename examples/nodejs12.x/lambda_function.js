"use strict";

exports.lambda_handler = async (event, context) => {

    if ("error" in event) {
        throw new Error(event.error);
    }

    console.log(JSON.stringify(event))
    console.log(JSON.stringify(context))

    return 'Hello World!';
}