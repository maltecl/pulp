const morphdom = require("morphdom")
const { defaultTags, ...otherTags } = require("./tags")
const { SD, FOR } = require("./types")

const { Assets } = require("./assets")


const { Routing } = require("./routing")

const morphdomHooks = (socket, handlers) => ({
    getNodeKey: function(node) {
        return node.id;
    },
    onBeforeNodeAdded: function(node) {
        return node;
    },
    onNodeAdded: function(node) {

        for (const { applyWhen, on, tag, handler }
            of handlers) {

            if (!applyWhen(node)) {
                continue
            }

            if (!node.hasAttribute(tag) && !node.hasAttribute(":" + tag)) {
                continue
            }


            let eventName = node.getAttribute(tag)
            if (eventName === null) {
                eventName = node.getAttribute(":" + tag)
            }

            node.addEventListener(on, (event) => {
                let payload = handler(event, eventName)
                if (payload === null) {
                    return
                }


                for (const attribute of node.attributes) {
                    if (attribute.name.startsWith(":value-")) {
                        const key = attribute.name.slice(":value-".length)
                        payload = {...payload, [key]: attribute.value.trim() }
                    }
                }

                // const values = node.getAttribute(":values")
                // if (values !== null && values.trim() !== "") {
                //     payload = {...payload, values: values }
                // }

                socket.ws.send(JSON.stringify(payload, null, 0))
            })
        }
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
        // note: all event-listeners should be removed automatically, as no one holds reference of the node 
        // see: https://stackoverflow.com/questions/12528049/if-a-dom-element-is-removed-are-its-listeners-also-removed-from-memory
    },
    onBeforeElChildrenUpdated: function(fromEl, toEl) {
        return true;
    },
    childrenOnly: false
})

class PulpSocket {

    constructor(mountID, { events, debug } = { events: [], debug: false }, ) {

        // const basePath = new URL(document.location.href).pathname

        let cachedSD = {}; // TODO: make this better somehow. it works for now 
        let cachedAssets = null
        let ws = null;
        let hasMounted = false



        mount = document.getElementById(mountID)

        ws = new WebSocket("ws://" + document.location.host + "/ws")

        Object.assign(globalThis, { PulpSocket: this })


        const hooks = morphdomHooks({ ws }, [...Object.values(defaultTags), ...events])


        ws.onopen = (it) => {
            if (debug) {
                console.log(`socket for ${mountID} connected: `, it)
            }
        }

        ws.onmessage = ({ data }) => {
            data.text()
                // .then(x => JSON.parse(x))
                .then(message => {

                    Object.assign(globalThis, { lastMessage: message })

                    const messageJSON = JSON.parse(message)
                    if (messageJSON.assets !== undefined) {
                        const { assets } = messageJSON
                        console.log(assets)
                        if (cachedAssets == null) {
                            cachedAssets = new Assets(assets)
                        } else {
                            cachedAssets = cachedAssets.patch(assets)
                        }

                        Object.assign(globalThis, { cachedAssets })

                        if (this.onassets !== undefined) {
                            this.onassets(cachedAssets.cache)
                        }
                    }


                    if (messageJSON.html === undefined) {
                        return
                    }

                    if (!hasMounted) {
                        cachedSD = new SD(messageJSON.html)
                        console.log(cachedSD)
                        Object.assign(globalThis, { cachedSD })


                        const temp = document.createElement("div")
                        temp.id = mountID
                        temp.innerHTML = cachedSD.render()
                        morphdom(mount, temp, hooks)

                        hasMounted = true
                        return
                    }

                    if (debug) {
                        console.log("got patch: ", message)
                    }

                    const patches = messageJSON.html

                    cachedSD = cachedSD.patch(patches)


                    Object.assign(globalThis, { cachedSD })

                    const temp = document.createElement("div")
                    temp.id = mountID
                    const lastRender = cachedSD.render()
                    Object.assign(globalThis, { lastRender })
                    temp.innerHTML = lastRender
                    morphdom(mount, temp, hooks)

                }).catch(console.error)
        }



        console.log("RMAKRER")
            // window.onpopstate = Routing.windowRouteChangedHandler({ ws })
    }
}

module.exports = { PulpSocket, tags: {...defaultTags, ...otherTags } }