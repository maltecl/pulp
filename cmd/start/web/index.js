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
                console.log("staticDynamicToString: " + staticDynamicToString(cachedSD))


                const temp = document.createElement("div")
                temp.id = "mount"
                temp.innerHTML = staticDynamicToString(cachedSD)
                morphdom(mount, temp, morphdomHooks)

                hasMounted = true
                return
            }

            console.log("got patch: ", message)

            const patches = JSON.parse(message)


            const d = Dynamics.patch(cachedSD.d, patches)
            cachedSD = {...cachedSD, d }


            // cachedSD = applyPatchesToCachedSD(cachedSD, patches)



            Object.assign(globalThis, { cachedSD })

            // console.log(cachedSD)

            const temp = document.createElement("div")
            temp.id = "mount"
            const lastRender = staticDynamicToString(cachedSD)
            Object.assign(globalThis, { lastRender })
            temp.innerHTML = lastRender
            morphdom(mount, temp, morphdomHooks)

        }).catch(console.error)
}






function staticDynamicToString({ s, d }) {
    let out = ""

    for (let i = 0; i < s.length; i++) {
        out += s[i]

        if (!d) {
            continue
        }

        if (i < d.length) {
            if (set(d[i].ds)) { // forTemplate
                const template = d[i]
                let forStr = ""

                template.ds.forEach(dynamic => {
                    forStr += staticDynamicToString({ s: template.s, d: dynamic })
                })
                out += forStr
            } else if (If.detect(d[i])) { // ifTemplate
                out += If.render(d[i])
            } else {
                out += d[i]
            }

        }
    }

    return out
}

const If = {
    render({ c, t, f }) {
        return staticDynamicToString(c ? t : f)
    },
    patch(old, patches) {
        // reset the dynamics, if condition changed (meaning, the other statics are rendered), and no new dynamics were provided
        // if (set(patches.c)) {
        //     const conditionChanged = old.c != patches.c
        //     d = set(patches.d) ? patches.d : []
        // }

        const ret = {
            c: set(patches.c) ? patches.c : old.c,
            t: { s: old.t.s, d: set(patches.t) ? Dynamics.patch(old.t.d, patches.t) : old.t.d },
            f: { s: old.f.s, d: set(patches.f) ? Dynamics.patch(old.f.d, patches.f) : old.f.d },
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
            if (If.detect(copy[key])) {

                console.log("go if: ", patches[key])

                copy[key] = If.patch(copy[key], patches[key])
            } else {
                copy[key] = patches[key]
            }
        })

        return copy
    },
    detect(it) {
        return Array.isArray(it)
    }
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