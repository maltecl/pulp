const Amigo = require("amgio_web")
const morphdom = require("morphdom")

const mount = document.getElementById("mount")

const morphdomHooks = {
    getNodeKey: function(node) {
        return node.id;
    },
    onBeforeNodeAdded: function(node) {
        return node;
    },
    onNodeAdded: function(node) {
        const maybeHandler = Amigo.addHandlersForElementNames[node.constructor.name]
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
        const maybeHandler = Amigo.removeHandlersForElementNames[node.constructor.name]
        maybeHandler && maybeHandler(node)
    },
    onBeforeElChildrenUpdated: function(fromEl, toEl) {
        return true;
    },
    childrenOnly: false
}


console.log("hello dasdsa")

Object.assign(globalThis, { Amigo })


ws = new WebSocket("ws://" + document.location.host + "/ws")


hasMounted = false


cachedSD = {};

ws.onmessage = ({
    data
}) => {
    data.text()
        .then(message => {

            Object.assign(globalThis, { lastMessage: JSON.parse(message) })


            if (!hasMounted) {

                cachedSD = JSON.parse(message)
                console.log(cachedSD)
                console.log("StaticDynamic.render: " + StaticDynamic.render(cachedSD))


                const temp = document.createElement("div")
                temp.id = "mount"
                temp.innerHTML = StaticDynamic.render(cachedSD)
                morphdom(mount, temp, morphdomHooks)

                hasMounted = true
                return
            }

            console.log("got patch: ", message)

            const patches = JSON.parse(message)

            cachedSD = StaticDynamic.patch(cachedSD, patches)


            // cachedSD = applyPatchesToCachedSD(cachedSD, patches)



            Object.assign(globalThis, { cachedSD })

            // console.log(cachedSD)

            const temp = document.createElement("div")
            temp.id = "mount"
            const lastRender = StaticDynamic.render(cachedSD)
            Object.assign(globalThis, { lastRender })
            temp.innerHTML = lastRender
            morphdom(mount, temp, morphdomHooks)

        }).catch(console.error)
}




const If = {
    render({ c, t, f }) {
        return StaticDynamic.render(c ? t : f)
    },
    patch(old, patches) {
        // what weird sorcery was this? even if this is/was needed, it's not programmed out completely
        // reset the dynamics, if condition changed (meaning, the other statics are rendered), and no new dynamics were provided
        // if (set(patches.c)) {
        //     const conditionChanged = old.c != patches.c
        //     d = set(patches.d) ? patches.d : []
        // }

        const ret = {
            c: set(patches.c) ? patches.c : old.c,
            t: set(patches.t) ? StaticDynamic.patch(old.t, patches.t) : old.t,
            f: set(patches.f) ? StaticDynamic.patch(old.f, patches.f) : old.f,
        }

        return ret
    },
    detect(it) {
        return set(it.c) || set(it.f) || set(it.t)
    }
}


const Dynamics = {
    render(list) {

    },
    patch(old, patches) {
        let copy = [...old]

        Object.keys(patches).forEach(key => {
            if (copy[key] !== null && copy[key] !== undefined) {
                if (Dynamics.detect(copy[key])) {
                    copy[key] = Dynamics.patch(copy[key], patches[key])
                    return
                }
                if (For.detect(copy[key])) {
                    copy[key] = For.patch(copy[key], patches[key])
                    return
                }
                if (If.detect(copy[key])) {
                    copy[key] = If.patch(copy[key], patches[key])
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


const StaticDynamic = {
    render({ s, d }) {
        let out = ""

        for (let i = 0; i < s.length; i++) {
            out += s[i]

            if (!d) {
                continue
            }

            if (i < d.length) {
                if (For.detect(d[i])) {
                    out += For.render(d[i])
                } else if (If.detect(d[i])) { // ifTemplate
                    out += If.render(d[i])
                } else {
                    out += d[i]
                }

            }
        }

        return out
    },
    patch({ s, d }, patches) {
        return { s, d: Dynamics.patch(d, patches) }
    },
    detect() {

    },
}



const For = {
    strategy: {
        append: 0,
    },

    render({ s, ds }) {
        let forStr = ""

        ds.forEach(dynamic => {
            forStr += StaticDynamic.render({ s, d: dynamic })
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
                case For.strategy.append:
                    return {...old, ds: Dynamics.patch([...old.ds, null], patches.ds) }
                default:
                    console.error("should not be reached in switch")
            }
        }



        return {...old, ds: Dynamics.patch(old.ds, patches.ds) }

    },
    detect(it) {
        return set(it.ds) // holds true for both the initial server push and the patches
    },
}

// function applyPatchesToCachedSD(cached, patches) /*new sd*/ {
//     const copy = {...cached }


//     Object.keys(patches).forEach(k => {
//         // if (set(copy.d[i].ds)) {
//         // console.log("FOR PATCH")
//         // } else 
//         if (If.detect(patches[k])) {
//             copy.d[k] = If.patch(copy.d[k], patches[k])
//         } else {
//             copy.d[k] = patches[k]
//         }
//     })

//     return copy
// }






const set = x => x !== undefined