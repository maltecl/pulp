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

            // console.log(cachedSD)

            const temp = document.createElement("div")
            temp.id = "mount"
            temp.innerHTML = staticDynamicToString(cachedSD)
            morphdom(mount, temp, morphdomHooks)

        }).catch(console.error)
}




// func (s StaticDynamic) String() string {
// 	res := strings.Builder{}

// 	for i := range s.Static {
// 		res.WriteString(s.Static[i])

// 		if ok := i < len(s.Dynamic); ok {
// 			res.WriteString(fmt.Sprint(s.Dynamic[i]))
// 		}
// 	}

// 	return res.String()
// }

function staticDynamicToString({ s, d }) {
    let out = ""

    for (let i = 0; i < s.length; i++) {
        out += s[i]

        if (i < d.length) {
            out += d[i]
        }
    }

    return out
}



// do this on the client side
// for k, patch := range map[int]interface{}(*patches) {
// 	lastRender.Dynamic[k] = patch
// }
function applyPatchesToCachedSD(cached, patches) /*new sd*/ {
    const copy = {...cached }


    Object.keys(patches).forEach(k => {
        copy.d[k] = patches[k]
    })

    return copy
}