import * as actions from './actions';
import reduce, { initialState } from './reducers';

describe('advertDetails reducer', () => {
    describe('GET_ADVERT_DETAILS_REQUEST', () => {
        it('should clear state of reducer', () => {
            expect(reduce(initialState, { type: actions.GET_ADVERT_DETAILS_REQUEST }))
                .toEqual(initialState);
        });
    });

    describe('GET_ADVERT_DETAILS_RESPONSE', () => {
        it('should store details entity from action to the reducer\'s state', () => {
            const detailsStub = { id: 2 };
            const expectedState = { ...initialState, ...detailsStub, loading: false };
            expect(reduce(initialState, { type: actions.GET_ADVERT_DETAILS_RESPONSE, details: detailsStub }))
                .toEqual(expectedState);
        });
    });

    describe('SKYCOIN_PRICE_RESPONSE', () => {
        it('should save received price to the reducer\'s state', () => {
            const priceStub = 123;
            const expectedState = { ...initialState, price: priceStub, loading: false };
            expect(reduce(initialState,
                { type: actions.SKYCOIN_PRICE_RESPONSE, price: priceStub }))
                .toEqual(expectedState);
        });
    });
});
