import React from 'react';
import { Field, reduxForm, Form } from 'redux-form';
import { Box } from 'grid-styled';

import { FormInput, FormMessage, } from 'components/layout/Form';
import { Submit } from 'components/layout/Button';
import { required, minLength } from 'validation/rules';

const r = required();
const minLength8 = minLength(8);

const passwordsMatch = (value, allValues) => allValues.newPassword && value !== allValues.newPassword ? 'Passwords does not match' : undefined;

export default reduxForm({ form: 'changePasswordSettingsForm' })(
    class extends React.Component {
        render() {
            const {
                handleSubmit,
                pristine,
                submitting,

                showSuccessMessage,
            } = this.props;

            return (
                <Form onSubmit={handleSubmit}>
                    <Box width={[1, 1, 1 / 2]}>
                        {showSuccessMessage && <FormMessage>Settings updated</FormMessage>}
                        <Field
                            name="oldPassword"
                            component={FormInput}
                            type="password"
                            label="Old password"
                            placeholder="Old password"

                            validate={[r]}
                            isRequired
                        />

                        <Field
                            name="newPassword"
                            component={FormInput}
                            type="password"
                            label="New password"
                            placeholder="New password"

                            isRequired
                            description="8 characters or more"
                            validate={[r, minLength8]}
                        />

                        <Field
                            name="confirmPassword"
                            component={FormInput}
                            type="password"
                            label="Confirm new password"
                            placeholder="Confirm new password"

                            validate={[r, passwordsMatch]}
                            isRequired
                        />

                        <Submit disabled={pristine || submitting} showSpinner={submitting} text="Save" primary />
                    </Box>
                </Form>
            )
        }
    });
