import React from 'react';
import { connect } from 'react-redux';
import { Link } from 'react-router-dom';
import { SubmissionError } from 'redux-form';
import { Helmet } from 'react-helmet';
import { Box } from 'grid-styled';

import { getPageTitle } from 'utils';
import LoginForm from './LoginForm';
import { login } from './actions';
import Container from '../../layout/Container';
import { U, H2, H3 } from '../../layout/Text';

class Login extends React.Component {
    onSubmit = (user) => {
        const { loginUser, location } = this.props;
        const fromLocation = location && location.state && location.state.from;

        return loginUser(user, fromLocation)
            .catch(err => {
                throw new SubmissionError(err)
            });
    }

    render() {
        return (
            <Container flex='1 0 auto' flexDirection="column" py={5}>
                <Helmet><title>{getPageTitle('Login')}</title></Helmet>
                <H2>Login</H2>
                <H3>Don't have an account? <Link to="/register"><U>Register</U></Link></H3>
                <Box mt={2}>
                    <LoginForm onSubmit={this.onSubmit} />
                </Box>
            </Container>
        );
    }
}

export default connect(null, { loginUser: login })(Login)
