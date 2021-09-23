class Assets {

    constructor(obj) {
        this.cache = obj
    }


    patch(patches) {
        let newAssets = {...this.cache }

        for (const key in patches) {
            if (patches[key] === null) { // this element should be deleted
                delete newAssets[key]
            } else { // new or old element. overwrite it
                newAssets[key] = patches[key]
            }
        }

        return new Assets(newAssets)
    }

}


module.exports = { Assets }