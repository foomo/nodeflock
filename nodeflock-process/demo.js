var components =  {
    Demo: {
        Simple : function(name) {
            return "simple name " + name;
        },
    },
    Complex : function(size) {
        return {
            size: size,
            date: new Date()
        }
    }
}

return components;

