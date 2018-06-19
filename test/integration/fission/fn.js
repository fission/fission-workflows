module.exports = async (context) => {
    const body = JSON.stringify(context.request.body);
    const headers = JSON.stringify(context.request.headers);
    return {
        status: 200,
        body: `body:${body}\nheaders:${headers}`
    };
}