import React from 'react';
import PropTypes from 'prop-types';
import { Field } from 'redux-form';

import { required } from 'validation/rules';
import { FormTextArea, FormDropdown, FormInput, FormGroup } from 'components/layout/Form';

const placeHolder = `Example: I can meet in the Starbucks on Main St,\nin McDonalds on Broad St, or anywhere in the "X" shopping district.\nI can meet anytime between 1-4pm and my minimum trade is 1 SKY.'`;

const r = required();

const LocationFormGroup = ({ states, countries, showStates, city }) =>
    <FormGroup label={'Your location'}>
        <Field name="countryCode" component={FormDropdown} options={countries} label={'Country'} isRequired validate={[r]} />
        {showStates &&
            <Field name="stateCode" component={FormDropdown} options={states} label={'State'} isRequired validate={[r]} />
        }
        <Field name="city" component={FormInput} label={'City'} isRequired validate={[r]}/>
        <Field name="postalCode" component={FormInput} label={'Postal code (required for most countries)'} isRequired validate={[r]} />
        <Field name="additionalInfo" component={FormTextArea} label={'Additional information (optional)'} tip={'Up to 3,000 characters'} placeholder={placeHolder} />
    </FormGroup>;

LocationFormGroup.propTypes = {
    states: PropTypes.arrayOf(PropTypes.shape({
        value: PropTypes.any.isRequired,
        text: PropTypes.string.isRequired,
    })),
    countries: PropTypes.arrayOf(PropTypes.shape({
        value: PropTypes.any.isRequired,
        text: PropTypes.string.isRequired,
    })),
    showStates: PropTypes.bool,
}

export default LocationFormGroup;
