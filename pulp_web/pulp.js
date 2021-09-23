const morphdom = require("morphdom")
const { defaultEvents, ...otherEvents } = require("./events")
const { SD, FOR } = require("./types")

const { Assets } = require("./assets")



const morphdomHooks = (socket, handlers, userHooks) => ({
    getNodeKey: function (node) {
        return node.id;
    },
    onBeforeNodeAdded: function (node) {
        return node;
    },
    onNodeAdded: function (node) {

        userHooks.onNodeAdded && userHooks.onNodeAdded(node)

        for (const { applyWhen, on, event, handler }
            of handlers) {

            if (!applyWhen(node)) {
                continue
            }

            if (!node.hasAttribute(event) && !node.hasAttribute(":" + event)) {
                continue
            }


            let eventName = node.getAttribute(event)
            if (eventName === null) {
                eventName = node.getAttribute(":" + event)
            }

            node.addEventListener(on, (event) => {
                let payload = handler(event, eventName)
                if (payload === null) {
                    return
                }


                for (const attribute of node.attributes) {
                    if (attribute.name.startsWith(":value-")) {
                        const key = attribute.name.slice(":value-".length)
                        payload = { ...payload, [key]: attribute.value.trim() }
                    }
                }

                socket.ws.send(JSON.stringify(payload, null, 0))
            })
        }

    },
    onBeforeElUpdated: function (fromEl, toEl) {
        return true;
    },
    onElUpdated: function (el) {

    },
    onBeforeNodeDiscarded: function (node) {
        return true;
    },
    onNodeDiscarded: function (node) {
        // note: all event-listeners should be removed automatically, as no one holds reference of the node 
        // see: https://stackoverflow.com/questions/12528049/if-a-dom-element-is-removed-are-its-listeners-also-removed-from-memory
    },
    onBeforeElChildrenUpdated: function (fromEl, toEl) {
        return true;
    },
    childrenOnly: false
})

class PulpSocket {

    constructor(mountID, wsPath, config) {
        const events = config.events || []
        const debug = config.debug || false
        const hooks = config.hooks || {}

        this.lastRoute = null
        let cachedSD = null;
        let cachedAssets = null

        const mount = document.getElementById(mountID)

        if (!wsPath.startsWith("/")) {
            wsPath = "/" + wsPath
        }

        this.ws = new WebSocket(new URL(wsPath, "ws://" + document.location.host).href)

        Object.assign(globalThis, { PulpSocket: this })

        const mHooks = morphdomHooks({ ws: this.ws }, [...Object.values(defaultEvents), ...events], hooks)


        this.ws.onopen = (it) => {
            debug && console.log(`socket for ${mountID} connected: `, it)
        }

        this.ws.onmessage = ({ data }) => {
            data.text()
                .then(x => [JSON.parse(x), x])
                .then(([messageJSON, raw]) => {

                    debug && console.log("got patch: ", raw, messageJSON)

                    if (messageJSON.assets !== undefined) {
                        const { assets } = messageJSON
                        debug && console.log(assets)
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
                    morphdom(mount, temp, mHooks)

                }).catch(console.error)
        }


        const self = this
        window.addEventListener("popstate", (e) => {
            self.ws.send(JSON.stringify({ from: this.lastRoute === null ? "" : this.lastRoute, to: new URL(document.location.href).pathname }, null, 0))
        })
    }
}


Object.assign(globalThis, { Pulp: { PulpSocket, events: { ...defaultEvents, ...otherEvents } } })

module.exports = { PulpSocket, events: { ...defaultEvents, ...otherEvents } }