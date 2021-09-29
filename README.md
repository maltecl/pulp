# pulp

Pulp allows you to write dynamic web-applications entirely in go, by reacting to events on the server-side.


```go
func (c index) Render(pulp.Socket) (pulp.HTML, pulp.Assets) {
  return pulp.L(`
    {{ if c.showMessage }}
      <h2> {{ c.message }} </h2>
    {{ end }}

    <button :click="increment"> increment </button> 
    <span> you have pressed the button {{ c.counter }} times </span> 

    <input :input="input-changed" value={{ c.inputValue }} />

    {{ for _, user := range users :key user.id}}
      ...
    {{ end }}
  `), nil
}
```

## Getting Started
The best way to start, is to copy one of the examples. Right now there are only two examples: one for when you want to use npm for including the client library, one for when you just want to include a bundled js file. When you don't use the already bundled file, you will need some tool for bundling the library and your js files. The example uses [browserify](https://browserify.org/). Install (globally) with `npm i -g browserify`. Run `GOBIN=<target directory> go install github.com/maltecl/pulp/cmd/gen@latest` to install the tool that will generate go code from your templates. Make sure `GOBIN` is in your PATH. 


Now, run the `run.sh` script and open the url in your browser. 

Pulp is built so, that you can integrate it in your existing app. 
There are 4 steps:
- You need a struct that implements `LiveComponent`
- expose it's websocket via `LiveSocket(newComponent func() LiveComponent) http.HandlerFunc`
- have an HTML Element with an ID
- use `new PulpSocket("<that ID>", "/livesocket")` to connect to the live-socket and mount it at that ID



## Livecycle
Pulp roughly uses the same methods as Phoenix LiveView, I do __not__ claim to have invented the mechanism.

Upon mount, the template you wrote will be rendered and sent to the client. From now on, the client uses the same websocket connection to send all the events, that should be reacted to, to the server. The server will then re-render the template, compare the old render with the new render and create patches from that. Those are then sent across the wire back to the client and will be (efficiently) patched into the dom.


In code, the lifecycle of the app is represented with the methods of the LiveComponent interface:
```go
type LiveComponent interface {
	Mount(Socket)
	Render(socket Socket) (HTML, Assets)
	HandleEvent(Event, Socket)
}
```
`Mount` is called on mount.

`Render` is called, once, immediately after Mount was called and after that, whenever the `socket.Update()` or `socket.Redirect()` methods are called.

`HandleEvent` is called, whenever the client sent a pulp-event. Call `socket.Update()` from inside `HandleEvent()` when you are done handling the event to reflect the changes in the client.


## Pulp Events
Pulp events are those things that start with ":". Because of a lack of time, only three (`:click`, `:input`, `:key-submit`) of those are so far implemented and pulled in by default. See [pulp_web/events.js](https://github.com/maltecl/pulp/blob/master/pulp_web/events.js) for how you would go and implement your own. You can tell pulp to use your own events in addition like this:
```js
const socket = new PulpSocket("mount", "/ws", {
    events: [
        ... your events ...
    ]
})
```
Template code:
```handlebars
 <input :input="<event name>" />
 ```

 where `event name` is the name, that is then passed to `HandleEvent`. Along with that are passed are all values of the attributes that start with `:value-`: 
```handlebars
 <input :input="<event name>" :value-some-value="you could use a dynamic value here" value={{ c.message }}/>
 ```
`:input` will also send the standard HTML `value`-attribute.
`HandleEvent` could handle the `:input`-event like so:

 ```go
func (c *index) HandleEvent(event pulp.Event, socket pulp.Socket) {
	e := event.(pulp.UserEvent)

	switch e.Name {
	case "<event name>":
    fmt.Println(e.Data["some-value"].(string))
    c.message = e.Data["value"].(string)
		socket.Update()
	}
}
 ```

## Assets
You can use the same websocket connection, that is used for sending the events/patches back and forth for sending values, which should not appear in the markup. Those values are returned from `Render` with the second return value:


```go
func (c index) Render(pulp.Socket) (pulp.HTML, pulp.Assets) {
	return pulp.L(`markup...`), pulp.Assets{
      "intVal": 10,
      "stringVal": "hello world",
  }
}
```
Pulp will also only upon mount send all of those, after that it will just send the ones that have changed.
In the client you receive the values like so:
```js
const socket = new PulpSocket("mount", "/ws", {})
socket.onassets(({ intVal, stringVal }) => {
  console.log(intVal, stringVal)
})
```

## Template Language

For fast renders and diffs, pulp uses it's own template language. This can be off-putting at first, but because the language is compiled to go code, it can directly reference surounding go code, variable values do not need to be passed in some context, like with other template languages.

The template language can be used __anywhere__ in your go code, not just in the `Render` method, as long as it is inside a `string` wrapped in `pulp.L`:
```go
func f(value int) pulp.HTML {
  return pulp.L("value: {{ value }}")
}
```

Dynamic values are passed in between two curly braces:

```handlebars
<span> {{ dynamicValue }} </span>
```

If expressions look like this:
```handlebars
{{ if dynamicValue > 10 }}
  <span>  {{ dynamicValue }} </span>
{{ else }}
  <h3> too bad </h3>
{{ end }}
```
The `else`-case is optional. Note, that the `dynamicValue > 10` is just standard go code and will be copied as is into the compiled source. This expression can be as complicated as any go expression with one exception: binding variables like in `if err := ...; err != nil` is not yet possible.


For loops on the other hand can do this:
```handlebars
{{ for i, user := range users :key user.id}}
  <li> {{ i }} - {{ renderUser(user) }} </li>
{{ end }}
```

The code before `:key` is copied into the compiled source, just like with the `if`-expression. The expression after `:key` is used as a key for the body of the `for`-loop. The key must be of type `string` and __must__ be specified. The mechanism used here is similar to the one react uses and makes for much smaller patches and more efficient patching. As in react (? not sure about the current state) using the index , of an element as the key, may result in weird behaviour. 





## Why does this exist?
As far as I am aware of, there are currently three other projects (I will link them here), that do the LiveView for go thing. I wrote my own version, because I wanted it to be as simple as writing LiveView components and also because I wanted to learn the details.



## What's planned?
I wrote this, (partly) because I needed the end result. There are many ways to improve/optimise this project. I will adress those probably sometime, when I feel like I really need them and have a decent solution in mind.

Things I would really like to add:
- components -> right now, one component cannot render another component directly
- a more complete router
- ? smaller patches using [json-path](https://jsonpath.com/) (for now it's okay though)

If you have anything you want to add or a question in general, let me know.









