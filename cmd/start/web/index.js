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

        console.log(node)

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

            Object.assign(globalThis, { lastMessage: message })


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

            cachedSD = applyPatchesToCachedSD(cachedSD, patches)


            Object.assign(globalThis, { cachedSD })

            // console.log(cachedSD)

            const temp = document.createElement("div")
            temp.id = "mount"
            temp.innerHTML = staticDynamicToString(cachedSD)
            morphdom(mount, temp, morphdomHooks)

        }).catch(console.error)
}






function staticDynamicToString({ s, d }) {
    let out = ""

    for (let i = 0; i < s.length; i++) {
        out += s[i]

        if (i < d.length) {


            if (d[i].c !== undefined) { // ifTemplate
                const template = d[i]
                const { c, t, f } = template

                let ifStr = ""
                if (c) {
                    ifStr = staticDynamicToString({ s: t, d: set(template.d) ? template.d : [] })
                } else {
                    ifStr = staticDynamicToString({ s: f, d: set(template.d) ? template.d : [] })
                }
                out += ifStr
            } else {
                out += d[i]
            }

        }
    }

    return out
}




function applyPatchesToCachedSD(cached, patches) /*new sd*/ {
    const copy = {...cached }


    Object.keys(patches).forEach(k => {
        if (isIf(patches[k])) {
            copy.d[k] = patchIf(copy.d[k], patches[k])
        } else {
            copy.d[k] = patches[k]
        }
    })

    return copy
}



const set = x => x !== undefined

function isIf(d) {
    return set(d.c) || set(d.t) || set(d.f) || set(d.d)
}

function patchIf(old, new_) {

    let d = set(new_.d) ? new_.d : old.d

    // reset the dynamics, if condition changed (meaning, the other statics are rendered), and no new dynamics were provided
    if (set(new_.c)) {
        const conditionChanged = old.c != new_.c
        d = set(new_.d) ? new_.d : []
    }


    const ret = {
        c: set(new_.c) ? new_.c : old.c,
        t: set(new_.t) ? new_.t : old.t,
        f: set(new_.f) ? new_.f : old.f,
        d: d,
    }

    Object.assign(globalThis, { lastPatch: ret })

    return ret

}