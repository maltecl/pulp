const morphdom = require("morphdom")




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
    static CLICK = "pulp-click"
    static INPUT = "pulp-input"
    static VALUES = "pulp-value"
    static SUBMIT = "pulp-submit"


    static addHandlersForElementNames = socket => ({
        "HTMLButtonElement": (node) => Pulp.addHandler(socket, node, Pulp.CLICK, "click"),
        "HTMLInputElement": (node) => Pulp.addHandler(socket, node, Pulp.INPUT, "input", (node, e) => (["value", node.value])),
    })

    static removeHandlersForElementNames = socket => ({
        "HTMLButtonElement": (node) => Pulp.addHandler(socket, node, Pulp.CLICK, "click"), // TODO: this should be removeHandler
        "HTMLInputElement": (node) => Pulp.addHandler(socket, node, Pulp.INPUT, "input"),
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


            socket.ws.send(JSON.stringify(payload, null, 0))
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




    static If = {
        render({ c, t, f }) {
            return Pulp.StaticDynamic.render(c ? t : f)
        },
        patch(old, patches) {
            const ret = {
                c: set(patches.c) ? patches.c : old.c,
                t: set(patches.t) ? Pulp.StaticDynamic.patch(old.t, patches.t) : old.t,
                f: set(patches.f) ? Pulp.StaticDynamic.patch(old.f, patches.f) : old.f,
            }

            return ret
        },
        detect(it) {
            return set(it.c) || set(it.f) || set(it.t)
        }
    }

    static Dynamics = {
        render(list) {

        },
        patch(old, patches) {
            let copy = [...old]

            Object.keys(patches).forEach(key => {
                if (copy[key] !== null && copy[key] !== undefined) {
                    if (Pulp.Dynamics.detect(copy[key])) {
                        copy[key] = Pulp.Dynamics.patch(copy[key], patches[key])
                        return
                    }
                    if (Pulp.For.detect(copy[key])) {
                        copy[key] = Pulp.For.patch(copy[key], patches[key])
                        return
                    }
                    if (Pulp.If.detect(copy[key])) {
                        copy[key] = Pulp.If.patch(copy[key], patches[key])
                        return
                    }
                }


                copy[key] = patches[key]
            })

            return copy
        },
        detect(it) {
            return Array.isArray(it)
        }
    }


    static StaticDynamic = {
        render({ s, d }) {
            let out = ""

            for (let i = 0; i < s.length; i++) {
                out += s[i]

                if (!d) {
                    continue
                }

                if (i < d.length) {
                    if (Pulp.StaticDynamic.detect(d[i])) {
                        out += Pulp.StaticDynamic.render(d[i])
                    } else if (Pulp.For.detect(d[i])) {
                        out += Pulp.For.render(d[i])
                    } else if (Pulp.If.detect(d[i])) { // ifTemplate
                        out += Pulp.If.render(d[i])
                    } else {
                        out += d[i]
                    }

                }
            }

            return out
        },
        patch({ s, d }, patches) {
            return { s, d: Pulp.Dynamics.patch(d, patches) }
        },
        detect(it) {
            return set(it.s) && set(it.d)
        },
    }


    static For = {
        strategy: {
            append: 0,
        },

        render({ s, ds }) {
            let forStr = ""

            ds.forEach(dynamic => {
                forStr += Pulp.StaticDynamic.render({ s, d: dynamic })
            })

            return forStr
        },
        patch(old, patches) {
            console.log("OLD ", old, " PATCHES ", patches)



            const maxKey = Object.keys(patches.ds).map(k => parseInt(k)).reduce(Math.max, -1)
            const shouldResize = maxKey >= old.ds.length
            console.log(maxKey, shouldResize)


            if (shouldResize) {
                switch (old.strategy) {
                    case Pulp.For.strategy.append:
                        return {...old, ds: Pulp.Dynamics.patch([...old.ds, null], patches.ds) }
                    default:
                        console.error("should not be reached in switch")
                }
            }



            return {...old, ds: Pulp.Dynamics.patch(old.ds, patches.ds) }

        },
        detect(it) {
            return set(it.ds) // holds true for both the initial server push and the patches
        },
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

                    Object.assign(globalThis, { lastMessage: JSON.parse(message) })


                    if (!hasMounted) {

                        cachedSD = JSON.parse(message)
                        console.log(cachedSD)
                        console.log("Pulp.StaticDynamic.render: " + Pulp.StaticDynamic.render(cachedSD))


                        const temp = document.createElement("div")
                        temp.id = "mount"
                        temp.innerHTML = Pulp.StaticDynamic.render(cachedSD)
                        morphdom(mount, temp, morphdomHooks({ ws }))

                        hasMounted = true
                        return
                    }

                    console.log("got patch: ", message)

                    const patches = JSON.parse(message)

                    cachedSD = Pulp.StaticDynamic.patch(cachedSD, patches)


                    Object.assign(globalThis, { cachedSD })

                    const temp = document.createElement("div")
                    temp.id = "mount"
                    const lastRender = Pulp.StaticDynamic.render(cachedSD)
                    Object.assign(globalThis, { lastRender })
                    temp.innerHTML = lastRender
                    morphdom(mount, temp, morphdomHooks({ ws }))

                }).catch(console.error)
        }

    }
}


module.exports = { PulpSocket, Pulp }