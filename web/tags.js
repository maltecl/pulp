// tags are the events pulp will pick up on and send via the wire
// they might as well be called events; I named them "tags" to differentiate between them and normal HTML events
// the field "description" is not actually used yet



const inputTag = {
    description: "in a text-input, whenever _any_ text is entered, fire off an event, including the standard HTML-value attributes value",
    applyWhen(node) {
        console.log("applying inputTag: ", ["HTMLInputElement", "HTMLTextAreaElement"].includes(node.constructor.name))

        return ["HTMLInputElement", "HTMLTextAreaElement"].includes(node.constructor.name)
    },
    on: "input",
    tag: "input",
    handler(e, name) {
        return { name, value: e.target.value }
    },
}


const clickTag = {
    description: "on a button or anchor tag, when clicked, fire off an event",
    applyWhen(node) {
        return ["HTMLButtonElement"].includes(node.constructor.name)
    },
    on: "click",
    tag: "click",
    handler(e, name) {
        return { name, marker: "dummy marker" }
    },
}

const keySubmitTag = {
    description: "in a text-input, when enter is entered, fire off an event",
    applyWhen(node) {
        return ["HTMLInputElement", "HTMLTextAreaElement"].includes(node.constructor.name)
    },
    on: "keydown", // uses the "keydonw" HTML Event
    tag: "key-submit", // is tagged with "key-submit". in the source code it looks like this: ":key-submit=<name>"
    handler(e, name) {
        if (e.keyCode !== 13) {
            return null // reject the event. Payload is not sent
        }
        e.preventDefault()
        return { name }
    },
}


module.exports = {
    defaultTags: {
        inputTag,
        clickTag,
    },
    keySubmitTag,
}