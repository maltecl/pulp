const set = x => x !== undefined


function classify(it) {

    if (SD.detect(it)) {
        return new SD(it)
    }

    if (IF.detect(it)) {
        return new IF(it)
    }

    if (FOR.detect(it)) {
        return new FOR(it)
    }


    return it

}
class SD {
    static detect = (it) => set(it.s) && set(it.d)

    static render({ s, d }) {
        let out = ""

        for (let i = 0; i < s.length; i++) {
            out += s[i]

            // if (!d) {
            //     continue
            // }

            if (i < d.length) {
                out += set(d[i].render) ? d[i].render() : d[i]
            }
        }

        return out
    }

    static patchListOfDynamics(list, patches) {
        let copy = [...list]

        Object.keys(patches).forEach(key => {
            if (copy[key] !== null && copy[key] !== undefined) {
                if (set(copy[key].patch)) {
                    copy[key] = copy[key].patch(patches[key])
                    return
                }
            }


            copy[key] = patches[key]
        })

        return copy
    }

    constructor({ s, d }) {
        this.s = s
        this.d = d.map(classify)
    }

    render() {
        return SD.render(this)
    }

    patch(patches) {
        return new SD({ s: this.s, d: SD.patchListOfDynamics(this.d, patches) })
    }
}


class IF {
    static detect = (it) => set(it.c) || set(it.f) || set(it.t)

    constructor({ c, t, f }) {
        this.c = c
        this.t = new SD(t)
        this.f = new SD(f)
    }

    render() { return SD.render(this.c ? this.t : this.f) }

    patch(patches) {
        return new IF({
            c: set(patches.c) ? patches.c : this.c,
            t: set(patches.t) ? new SD(this.t).patch(patches.t) : this.t,
            f: set(patches.f) ? new SD(this.f).patch(patches.f) : this.f,
        })
    }
}


class FOR {
    static strategy = {
        append: 0,
    }

    static detect = (it) => set(it.ds) /* holds true for both the initial server push and the patches*/

    constructor({ strategy, ds, s }) {
        this.strategy = strategy
        this.s = s
        this.ds = ds.map(x => x.map(classify))
    }

    render() {
        let out = ""

        this.ds.forEach(dynamic => {
            out += SD.render({ s: this.s, d: dynamic })
        })

        return out
    }

    patch(patches) {
        const maxKey = Object.keys(patches.ds).map(k => parseInt(k)).reduce(Math.max, -1)
        const shouldResize = maxKey >= this.ds.length
        console.log(maxKey, shouldResize)


        if (shouldResize) {
            switch (this.strategy) {
                case FOR.strategy.append:
                    return new FOR({...this, ds: SD.patchListOfDynamics([...this.ds, null], patches.ds) })
                default:
                    console.error("should not be reached in switch")
            }
        }

        return new FOR({...this, ds: SD.patchListOfDynamics(this.ds, patches.ds) })

    }
}


module.exports = { SD, FOR, IF }