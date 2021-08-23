module.exports = class Pulp {
    static CLICK = "pulp-click"
    static INPUT = "pulp-input"
    static VALUES = "pulp-value"
    static SUBMIT = "pulp-submit"


    static addHandlersForElementNames = {
        "HTMLButtonElement": (node) => Pulp.addHandler(node, Pulp.CLICK, "click"),
        "HTMLInputElement": (node) => Pulp.addHandler(node, Pulp.INPUT, "input", (node, e) => (["value", node.value])),
    }

    static removeHandlersForElementNames = {
        "HTMLButtonElement": (node) => Pulp.addHandler(node, Pulp.CLICK, "click"),
        "HTMLInputElement": (node) => Pulp.addHandler(node, Pulp.INPUT, "input"),
    }

    static handlerForNode(node, pulpAttr, includeValues) {
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

            ws.send(JSON.stringify(payload, null, 0))
        }
    }

    static addHandler(node, pulpAttr, eventType, ...includeValues) {
        if (node.hasAttribute(pulpAttr)) {
            node.addEventListener(eventType, Pulp.handlerForNode(node, pulpAttr, includeValues))
        }
    }

    static removeHandler(node, pulpAttr, eventType) {
        if (node.hasAttribute(pulpAttr)) {
            node.removeEventListener(eventType, Pulp.handlerForNode(node, pulpAttr))
        }
    }
}