// const Amigo = require("./amigo")
// const morphdom = require("morphdom")

// const mount = document.getElementById("mount")

// const morphdomHooks = {
//     getNodeKey: function(node) {
//         return node.id;
//     },
//     onBeforeNodeAdded: function(node) {
//         return node;
//     },
//     onNodeAdded: function(node) {
//         const maybeHandler = Amigo.addHandlersForElementNames[node.constructor.name]
//         maybeHandler && maybeHandler(node)
//     },
//     onBeforeElUpdated: function(fromEl, toEl) {
//         return true;
//     },
//     onElUpdated: function(el) {

//     },
//     onBeforeNodeDiscarded: function(node) {
//         return true;
//     },
//     onNodeDiscarded: function(node) {
//         const maybeHandler = Amigo.removeHandlersForElementNames[node.constructor.name]
//         maybeHandler && maybeHandler(node)
//     },
//     onBeforeElChildrenUpdated: function(fromEl, toEl) {
//         return true;
//     },
//     childrenOnly: false
// }


// console.log("hello dasdsa")

// ws = new WebSocket("ws://" + document.location.host + "/ws")



// ws.onmessage = ({
//     data
// }) => {
//     data.text()
//         .then(patch => {
//             console.log("got patch")
//             console.log(patch)

//             const temp = document.createElement("div")
//             temp.id = "mount"
//             temp.innerHTML = patch
//             morphdom(mount, temp, morphdomHooks)

//         }).catch(console.error)
// }