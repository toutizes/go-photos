import {expect,it} from 'vitest';
import {constrain,identity,pinch} from './gestures';
it('keeps the whole image centered at fit scale',()=>{expect(constrain({scale:1,x:100,y:-200},800,400,1000,1000)).toEqual(identity);});
it('allows pan only across the scaled image bounds',()=>{expect(constrain({scale:2,x:900,y:900},800,400,1600,800)).toEqual({scale:2,x:400,y:200});});
it('keeps the point beneath the pinch center anchored',()=>{expect(pinch(identity,{x:100,y:50},{x:100,y:50},2)).toEqual({scale:2,x:-100,y:-50});});
it('limits pinch scale without overshooting translation',()=>{expect(pinch({scale:4,x:0,y:0},{x:100,y:0},{x:100,y:0},2)).toEqual({scale:5,x:-25,y:0});});
