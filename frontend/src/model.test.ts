import { describe, expect, it } from 'vitest';
import { readRoute, routeHash, sortedPhotos, photoPath, stereoOf, type Photo } from './model';
const photo = (Id: number, Its: number): Photo => ({Id,Its,Ad:'2025/été #1',In:'photo A.JPG',W:2000,H:1000,Fts:0,Kwd:[]});
describe('API and navigation compatibility', () => {
  it('round-trips French searches and full-screen photo links', () => {
    const route = {page:'photos' as const,q:'été, famille & amis',id:42,full:true};
    expect(readRoute(routeHash(route))).toEqual(route);
    expect(readRoute('#/photos?id=bad').id).toBeUndefined();
  });
  it('sorts albums chronologically and other searches newest first without mutating cache', () => {
    const input = [photo(1,10),photo(2,20)];
    expect(sortedPhotos(input,'été').map(p=>p.Id)).toEqual([2,1]);
    expect(sortedPhotos(input,'album:2025/été').map(p=>p.Id)).toEqual([1,2]);
    expect(input[0].Id).toBe(1);
  });
  it('encodes individual path segments and preserves JPG filename case', () => {
    expect(photoPath(photo(1,10))).toBe('/db/midi/2025/%C3%A9t%C3%A9%20%231/photo%20A.JPG');
  });
  it('preserves server stereo offsets even when stereo keywords are present', () => {
    const p = {...photo(1,10),Kwd:['stéréo'],Stereo:{Dx:14,Dy:-8}};
    expect(stereoOf(p)).toEqual({Dx:14,Dy:-8});
  });
});
