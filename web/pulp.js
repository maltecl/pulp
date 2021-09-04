const morphdom = require("morphdom")
const { FOR, IF, SD } = require("./types")



const morphdomHooks = socket => ({
    getNodeKey: function(node) {
        return node.id;
    },
    onBeforeNodeAdded: function(node) {
        return node;
    },
    onNodeAdded: function(node) {
        console.log("added")
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





const set = x => x !== undefined


class PulpSocket {

    constructor(mountID) {

        let cachedSD = {}; // TODO: make this better somehow. it works for now 
        let ws = null;
        let hasMounted = false




        mount = document.getElementById(mountID)

        ws = new WebSocket("ws://" + document.location.host + "/ws")

        Object.assign(globalThis, { PulpSocket: this })


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
                        morphdom(mount, temp, morphdomHooks({ ws }))

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
                    morphdom(mount, temp, morphdomHooks({ ws }))

                }).catch(console.error)
        }

    }
}





module.exports = { PulpSocket, Pulp }