(function(){function r(e,n,t){function o(i,f){if(!n[i]){if(!e[i]){var c="function"==typeof require&&require;if(!f&&c)return c(i,!0);if(u)return u(i,!0);var a=new Error("Cannot find module '"+i+"'");throw a.code="MODULE_NOT_FOUND",a}var p=n[i]={exports:{}};e[i][0].call(p.exports,function(r){var n=e[i][1][r];return o(n||r)},p,p.exports,r,e,n,t)}return n[i].exports}for(var u="function"==typeof require&&require,i=0;i<t.length;i++)o(t[i]);return o}return r})()({1:[function(require,module,exports){
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
},{}]},{},[1]);
