const morphdom = require("morphdom")
const { FOR, IF, SD } = require("./types")


/*


const inputEvents = [{
    applyWhen(node) {
        return node.constructor.name in ["HTMLInputElement", "HTMLTextAreaElement"]
    },
    on: "keydown", // uses the "keydonw" HTML Event
    tag: "key-submit", // is tagged with "key-submit". in the source code it looks like this: ":key-submit=<name>"
    handler(e, name) {
        if (e.keyCode !== 13) {
            return null // reject the event. Payload is not sent
        }
        return { name }
    },
}]

const socket = new PulpSocket("mount", {
    events: [
        ...inputEvents,
    ],

})

*/



const morphdomHooks = (socket, handlers) => ({
    getNodeKey: function(node) {
        return node.id;
    },
    onBeforeNodeAdded: function(node) {
        return node;
    },
    onNodeAdded: function(node) {
        console.log("added")


        for (const { applyWhen, on, tag, handler }
            of handlers) {

            console.log("MARKER 2")
            if (!applyWhen(node)) {
                continue
            }
            console.log("MARKER 3")

            if (!node.hasAttribute(tag) && !node.hasAttribute(":" + tag)) {
                continue
            }

            console.log("MARKER 4")


            let eventName = node.getAttribute(tag)
            if (eventName === null) {
                eventName = node.getAttribute(":" + tag)
            }

            node.addEventListener(on, (event) => {
                let payload = handler(event, eventName)
                if (payload === null) {
                    return
                }

                payload = { type: payload.name }

                const value = node.getAttribute(Pulp.VALUES)
                if (value !== null && value.trim() !== "") {
                    payload = {...payload, value: value }
                }

                socket.ws.send(JSON.stringify(payload, null, 0))
            })
        }

        const maybeHandler = Pulp.addHandlersForElementNames(socket)[node.constructor.name]

        maybeHandler && maybeHandler(node)
    },
    onBeforeElUpdated: function(fromEl, toEl) {
        return true;
    },
    onElUpdated: function(el) {

    },
    onBeforeNodeDiscarded: function(node) {
        return true;
    },
    onNodeDiscarded: function(node) {
        const maybeHandler = Pulp.removeHandlersForElementNames(socket)[node.constructor.name]
        maybeHandler && maybeHandler(node)
    },
    onBeforeElChildrenUpdated: function(fromEl, toEl) {
        return true;
    },
    childrenOnly: false
})

// temp0.addEventListener("keydown", e => console.log(e))
class Pulp {

    static DEBUG = true

    static CLICK = ":click"
    static INPUT = ":input"
    static VALUES = ":value"
    static SUBMIT = "pulp-submit"


    static addHandlersForElementNames = socket => ({
        "HTMLButtonElement": (node) => Pulp.addHandler(socket, node, Pulp.CLICK, "click"),
        "HTMLInputElement": (node) => Pulp.addHandler(socket, node, Pulp.INPUT, "input", (node, e) => (["value", node.value])),
    })

    static removeHandlersForElementNames = socket => ({
        "HTMLButtonElement": (node) => Pulp.removeHandler(socket, node, Pulp.CLICK, "click"),
        "HTMLInputElement": (node) => Pulp.removeHandler(socket, node, Pulp.INPUT, "input"),
    })

    static handlerForNode(socket, node, pulpAttr, includeValues) {
        return (e) => {
            let payload = {
                type: node.getAttribute(pulpAttr),
            }



            includeValues && includeValues.map(iv => iv(node, e)).forEach((maybeKeyVal) => {
                if (!maybeKeyVal) {
                    return
                }

                const [key, value] = maybeKeyVal


                payload = {...payload, [key]: value }
            })

            const value = node.getAttribute(Pulp.VALUES)
            if (value !== null && value.trim().length !== 0) {
                payload = {...payload, value: value }
            }


            const str = JSON.stringify(payload, null, 0)

            if (Pulp.DEBUG) {
                console.log("payload: ", str)
            }


            socket.ws.send(str)
        }
    }

    static addHandler(socket, node, pulpAttr, eventType, ...includeValues) {
        if (node.hasAttribute(pulpAttr)) {
            node.addEventListener(eventType, Pulp.handlerForNode(socket, node, pulpAttr, includeValues))
        }
    }

    static removeHandler(socket, node, pulpAttr, eventType) {
        if (node.hasAttribute(pulpAttr)) {
            node.removeEventListener(eventType, Pulp.handlerForNode(socket, node, pulpAttr))
        }
    }
}







class PulpSocket {

    constructor(mountID, { events } = { events: [] }) {

        let cachedSD = {}; // TODO: make this better somehow. it works for now 
        let ws = null;
        let hasMounted = false



        mount = document.getElementById(mountID)

        ws = new WebSocket("ws://" + document.location.host + "/ws")

        Object.assign(globalThis, { PulpSocket: this })


        const hooks = morphdomHooks({ ws }, events)

        ws.onmessage = ({ data }) => {
            data.text()
                .then(message => {

                    Object.assign(globalThis, { lastMessage: message })


                    if (!hasMounted) {

                        cachedSD = new SD(JSON.parse(message))
                        console.log(cachedSD)
                        Object.assign(globalThis, { cachedSD })


                        const temp = document.createElement("div")
                        temp.id = "mount"
                        temp.innerHTML = cachedSD.render()
                        morphdom(mount, temp, hooks)

                        hasMounted = true
                        return
                    }

                    console.log("got patch: ", message)

                    const patches = JSON.parse(message)

                    cachedSD = cachedSD.patch(patches)


                    Object.assign(globalThis, { cachedSD })

                    const temp = document.createElement("div")
                    temp.id = "mount"
                    const lastRender = cachedSD.render()
                    Object.assign(globalThis, { lastRender })
                    temp.innerHTML = lastRender
                    morphdom(mount, temp, hooks)

                }).catch(console.error)
        }

    }
}





module.exports = { PulpSocket, Pulp }