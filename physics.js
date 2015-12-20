window.onload = function() {

    // Matter.js module aliases
    var Engine = Matter.Engine,
        World = Matter.World,
        Bodies = Matter.Bodies,
        Composite = Matter.Composite,
        Common = Matter.Common,
        Body = Matter.Body,
        Constraint = Matter.Constraint,
        Composites = Matter.Composites;

    // create a Matter.js engine
    engine = Engine.create(document.body);

    var bridge = Composites.stack(150, 300, 9, 1, 10, 10, function(x, y) {
        return Bodies.rectangle(x, y, 50, 20);
    });

    Composites.chain(bridge, 0.5, 0, -0.5, 0, {
        stiffness: 0.6
    });

    var stack = Composites.stack(200, 40, 6, 3, 0, 0, function(x, y) {
        return Bodies.polygon(x, y, Math.round(Common.random(1, 8)), Common.random(20, 40));
    });

    // create two boxes and a ground
    var boxA = Bodies.rectangle(400, 200, 80, 80);
    var boxB = Bodies.rectangle(450, 50, 80, 80);
    var ground = Bodies.rectangle(400, 610, 810, 60, {
        isStatic: true
    });

    // create mouse constraint
    var mouseConstraint = Matter.MouseConstraint.create(engine);
    World.add(engine.world, [mouseConstraint,
        bridge,
        Bodies.rectangle(80, 440, 120, 280, {
            isStatic: true
        }),
        Bodies.rectangle(720, 440, 120, 280, {
            isStatic: true
        }),
        Constraint.create({
            pointA: {
                x: 140,
                y: 300
            },
            bodyB: bridge.bodies[0],
            pointB: {
                x: -25,
                y: 0
            }
        }),
        Constraint.create({
            pointA: {
                x: 660,
                y: 300
            },
            bodyB: bridge.bodies[8],
            pointB: {
                x: 25,
                y: 0
            }
        }),
    ]);

    // Allow the users to drag around the physics objects
    engine.render.mouse = mouseConstraint.mouse;

    // Add all of the bodies to the world
    World.add(engine.world, [mouseConstraint, ground]);

    // run the engine
    Engine.run(engine);

    // Socket to listen on
    var sock = new WebSocket("ws://localhost:12345/");

    var count = 1;
    var bodies = [];

    // Users object is unique, polygon instead of circle
    var user = Bodies.polygon(400, 200, 5, 20);

    // Add users object to body array
    bodies.push(user);

    // Add all the bodies to the world
    World.add(engine.world, [user]);

    // Show FPS
    engine.render.options.showDebug = true;

    // Called when socket successfully connects
    sock.onopen = function(evt) {}

    // Add or remove objects from the world to match the server count
    function balance(newCount) {
        while (newCount > count) {
            var newBox = Bodies.circle(400, 50, 20);
            bodies.push(newBox);
            World.add(engine.world, [newBox]);
            count += 1;
        }
        while (newCount < count) {
            Composite.remove(engine.world, bodies.pop());
            count -= 1;
        }
    }

    // Called when the client receives a message from the server
    sock.onmessage = function(evt) {
        console.log(evt.data);
        var response = JSON.parse(evt.data);
        if (response.count) {
            balance(response.count);
        }
    }
}