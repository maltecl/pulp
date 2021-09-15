module.exports.Routing = class Routing {


    static windowRouteChangedHandler({ ws }) {
        return (x1, x2, x3) => {
            ws.send(JSON.stringify({ from: "", to: new URL(document.location.href).pathname }, null, 0))
            console.log(x1, x2, x3)
        }
    }


    static serverRouteChangedHandler() {

    }

}