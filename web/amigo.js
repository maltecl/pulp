module.exports = class Amigo {
    static CLICK = "amigo-click"
    static INPUT = "amigo-input"
    static VALUES = "amigo-value"

    static addHandlersForElementNames = {
        "HTMLButtonElement": (node) => Amigo.addHandler(node, Amigo.CLICK, "click"),
        "HTMLInputElement": (node) => Amigo.addHandler(node, Amigo.INPUT, "input", (node, e) => (["value", node.value])),
    }

    static removeHandlersForElementNames = {
        "HTMLButtonElement": (node) => Amigo.addHandler(node, Amigo.CLICK, "click"),
        "HTMLInputElement": (node) => Amigo.addHandler(node, Amigo.INPUT, "input"),
    }

    static handlerForNode(node, amigoAttr, includeValues) {
        return (e) => {
            let payload = {
                type: node.getAttribute(amigoAttr),
            }



            includeValues && includeValues.map(iv => iv(node, e)).forEach((maybeKeyVal) => {
                if (!maybeKeyVal) {
                    return
                }

                const [key, value] = maybeKeyVal


                payload = {...payload, [key]: value }
            })

            const value = node.getAttribute(Amigo.VALUES)
            if (value !== null && value.trim().length !== 0) {
                payload = {...payload, value: value }
            }

            ws.send(JSON.stringify(payload, null, 0))
        }
    }

    static addHandler(node, amigoAttr, eventType, ...includeValues) {
        if (node.hasAttribute(amigoAttr)) {
            node.addEventListener(eventType, Amigo.handlerForNode(node, amigoAttr, includeValues))
        }
    }

    static removeHandler(node, amigoAttr, eventType) {
        if (node.hasAttribute(amigoAttr)) {
            node.removeEventListener(eventType, Amigo.handlerForNode(node, amigoAttr))
        }
    }
}