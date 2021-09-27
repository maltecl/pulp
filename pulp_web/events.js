// pulp events are the events pulp will pick up on and send via the wire
// the field "description" is not actually used yet



const inputEvent = {
    description: "in a text-input, whenever _any_ text is entered, fire off an event, including the standard HTML-value attributes value",
    applyWhen(node) {
        return ["HTMLInputElement", "HTMLTextAreaElement"].includes(node.constructor.name)
    },
    on: "input",
    event: "input",
    handler(e, name) {
        return { name, value: e.target.value }
    },
}


const clickEvent = {
    description: "on a button or anchor tag, when clicked, fire off an event",
    applyWhen(node) {
        return ["HTMLButtonElement"].includes(node.constructor.name)
    },
    on: "click",
    event: "click",
    handler(e, name) {
        return { name }
    },
}

const keySubmitEvent = {
    description: "in a text-input, when enter is entered, fire off an event",
    applyWhen(node) {
        return ["HTMLInputElement", "HTMLTextAreaElement"].includes(node.constructor.name)
    },
    on: "keydown", // uses the "keydonw" HTML Event
    event: "key-submit", // is tagged with "key-submit". in the source code it looks like this: ":key-submit=<name>"
    handler(e, name) {
        console.log("MARKER2")
        const enterKeyCode = 13
        if (e.keyCode !== enterKeyCode) {
            return null // reject the event. Payload is not sent
        }
        e.preventDefault()
        return { name }
    },
}

module.exports = {
    defaultEvents: {
        inputTag: inputEvent,
        clickTag: clickEvent,
    },
    keySubmitTag: keySubmitEvent,
}