const morphdom = require("morphdom")
const { defaultTags, ...otherTags } = require("./tags")
const { SD, FOR } = require("./types")

const { Assets } = require("./assets")



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

    constructor(mountID, wsPath, { events, debug } = { events: [], debug: false }, ) {
        this.lastRoute = null


        let cachedSD = null;
        let cachedAssets = null



        const mount = document.getElementById(mountID)


        if (!wsPath.startsWith("/")) {
            wsPath = "/" + wsPath
        }

        this.ws = new WebSocket(new URL(wsPath, "ws://" + document.location.host).href)


        Object.assign(globalThis, { PulpSocket: this })


        const hooks = morphdomHooks({ ws: this.ws }, [...Object.values(defaultTags), ...events])


        this.ws.onopen = (it) => {
            if (debug) {
                console.log(`socket for ${mountID} connected: `, it)
            }
        }

        this.ws.onmessage = ({ data }) => {
            data.text()
                .then(x => JSON.parse(x))
                .then(messageJSON => {

                    if (debug) {
                        console.log("got patch: ", messageJSON)
                    }

                    if (messageJSON.assets !== undefined) {
                        const { assets } = messageJSON
                        console.log(assets)
                        if (cachedAssets == null) {
                            cachedAssets = new Assets(assets)
                        } else {
                            cachedAssets = cachedAssets.patch(assets)
                        }

                        Object.assign(globalThis, { cachedAssets })

                        const { route } = assets
                        history.pushState({}, null, route)
                        this.lastRoute = route


                        if (this.onassets !== undefined) {
                            this.onassets(cachedAssets.cache)
                        }
                    }


                    if (messageJSON.html !== undefined) {
                        if (cachedSD === null) { // has not mounted yet => no patching
                            cachedSD = new SD(messageJSON.html)
                        } else {
                            const patches = messageJSON.html
                            cachedSD = cachedSD.patch(patches)
                        }
                    }

                    const temp = document.createElement("div")
                    temp.id = mountID
                    temp.innerHTML = cachedSD.render()
                    morphdom(mount, temp, hooks)

                }).catch(console.error)
        }


        const self = this
        window.addEventListener("popstate", (e) => {
            console.log("pop: ", e)
            self.ws.send(JSON.stringify({ from: this.lastRoute === null ? "" : this.lastRoute, to: new URL(document.location.href).pathname }, null, 0))
        })
    }
}

module.exports = { PulpSocket, tags: {...defaultTags, ...otherTags } }